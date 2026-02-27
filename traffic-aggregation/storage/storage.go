package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"visual-network/traffic-aggregation/types"
)

// Store handles persistent JSON storage of network data
type Store struct {
	dataDir  string
	mutex    sync.Mutex
	maxFiles int // Keep only this many files per type
}

// NewStore creates a new data store
func NewStore(dataDir string, maxFiles int) (*Store, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	store := &Store{
		dataDir:  dataDir,
		maxFiles: maxFiles,
	}

	log.Printf("Storage initialized at: %s (keeping last %d files per type)", dataDir, maxFiles)
	return store, nil
}

// SaveTopologyUpdate persists a topology update to JSON file
func (s *Store) SaveTopologyUpdate(update types.TopologyUpdate) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	timestamp := update.Timestamp.Format("2006-01-02T15-04-05")
	filename := filepath.Join(s.dataDir, fmt.Sprintf("topology_%s.json", timestamp))

	data, err := json.MarshalIndent(update, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal topology update: %w", err)
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write topology file: %w", err)
	}

	log.Printf("Saved topology update: %s", filename)

	// Cleanup old files
	s.cleanupOldFiles("topology_")
	return nil
}

// SaveMetricsUpdate persists a metrics update to JSON file
func (s *Store) SaveMetricsUpdate(update types.MetricsUpdate) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	timestamp := update.Timestamp.Format("2006-01-02T15-04-05")
	filename := filepath.Join(s.dataDir, fmt.Sprintf("metrics_%s.json", timestamp))

	data, err := json.MarshalIndent(update, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics update: %w", err)
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	log.Printf("Saved metrics update: %s", filename)

	// Cleanup old files
	s.cleanupOldFiles("metrics_")
	return nil
}

// ListTopologyFiles returns a list of available topology snapshots
func (s *Store) ListTopologyFiles() ([]string, error) {
	return s.listFiles("topology_")
}

// ListMetricsFiles returns a list of available metrics snapshots
func (s *Store) ListMetricsFiles() ([]string, error) {
	return s.listFiles("metrics_")
}

// listFiles returns sorted list of files with given prefix
func (s *Store) listFiles(prefix string) ([]string, error) {
	files, err := ioutil.ReadDir(s.dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	var results []string
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), prefix) {
			results = append(results, filepath.Join(s.dataDir, file.Name()))
		}
	}

	// Sort by modification time, newest first
	sort.Slice(results, func(i, j int) bool {
		fi, _ := os.Stat(results[i])
		fj, _ := os.Stat(results[j])
		return fi.ModTime().After(fj.ModTime())
	})

	return results, nil
}

// LoadTopologyFile loads a topology snapshot from file
func (s *Store) LoadTopologyFile(filepath string) (*types.TopologyUpdate, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var update types.TopologyUpdate
	if err := json.Unmarshal(data, &update); err != nil {
		return nil, fmt.Errorf("failed to unmarshal topology: %w", err)
	}

	return &update, nil
}

// LoadMetricsFile loads a metrics snapshot from file
func (s *Store) LoadMetricsFile(filepath string) (*types.MetricsUpdate, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var update types.MetricsUpdate
	if err := json.Unmarshal(data, &update); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return &update, nil
}

// GetLatestTopology returns the most recent topology snapshot
func (s *Store) GetLatestTopology() (*types.TopologyUpdate, error) {
	files, err := s.ListTopologyFiles()
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no topology files found")
	}

	return s.LoadTopologyFile(files[0])
}

// GetLatestMetrics returns the most recent metrics snapshot
func (s *Store) GetLatestMetrics() (*types.MetricsUpdate, error) {
	files, err := s.ListMetricsFiles()
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no metrics files found")
	}

	return s.LoadMetricsFile(files[0])
}

// cleanupOldFiles removes old files, keeping only maxFiles
func (s *Store) cleanupOldFiles(prefix string) {
	files, err := ioutil.ReadDir(s.dataDir)
	if err != nil {
		log.Printf("Error reading directory for cleanup: %v", err)
		return
	}

	var matching []os.FileInfo
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), prefix) {
			matching = append(matching, file)
		}
	}

	// Sort by modification time, oldest first
	sort.Slice(matching, func(i, j int) bool {
		return matching[i].ModTime().Before(matching[j].ModTime())
	})

	// Delete old files if we exceed maxFiles
	if len(matching) > s.maxFiles {
		for i := 0; i < len(matching)-s.maxFiles; i++ {
			filepath := filepath.Join(s.dataDir, matching[i].Name())
			if err := os.Remove(filepath); err != nil {
				log.Printf("Failed to remove old file %s: %v", filepath, err)
			} else {
				log.Printf("Cleaned up old file: %s", matching[i].Name())
			}
		}
	}
}

// ExportSnapshot creates a complete snapshot combining topology and metrics
func (s *Store) ExportSnapshot() (*Snapshot, error) {
	topology, topoErr := s.GetLatestTopology()
	metrics, metricsErr := s.GetLatestMetrics()

	if topoErr != nil && metricsErr != nil {
		return nil, fmt.Errorf("no data available to export")
	}

	snapshot := &Snapshot{
		ExportTime: time.Now(),
		Topology:   topology,
		Metrics:    metrics,
	}

	return snapshot, nil
}

// Snapshot represents a combined topology and metrics snapshot
type Snapshot struct {
	ExportTime time.Time             `json:"export_time"`
	Topology   *types.TopologyUpdate `json:"topology,omitempty"`
	Metrics    *types.MetricsUpdate  `json:"metrics,omitempty"`
}

// SaveSnapshot saves a complete snapshot to file
func (s *Store) SaveSnapshot(snapshot *Snapshot) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	timestamp := snapshot.ExportTime.Format("2006-01-02T15-04-05")
	filename := filepath.Join(s.dataDir, fmt.Sprintf("snapshot_%s.json", timestamp))

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	log.Printf("Saved snapshot: %s", filename)

	// Cleanup old files
	s.cleanupOldFiles("snapshot_")
	return nil
}

// GetStorageStats returns statistics about stored data
func (s *Store) GetStorageStats() (StorageStats, error) {
	stats := StorageStats{
		Timestamp: time.Now(),
	}

	topoFiles, err := s.ListTopologyFiles()
	if err == nil {
		stats.TopologyCount = len(topoFiles)
	}

	metricsFiles, err := s.ListMetricsFiles()
	if err == nil {
		stats.MetricsCount = len(metricsFiles)
	}

	// Calculate total directory size
	err = filepath.Walk(s.dataDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			stats.TotalBytes += info.Size()
		}
		return nil
	})

	stats.StorageDir = s.dataDir
	return stats, nil
}

// StorageStats provides information about stored data
type StorageStats struct {
	Timestamp     time.Time `json:"timestamp"`
	StorageDir    string    `json:"storage_dir"`
	TopologyCount int       `json:"topology_count"`
	MetricsCount  int       `json:"metrics_count"`
	TotalBytes    int64     `json:"total_bytes"`
}
