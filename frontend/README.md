## Frontend Installation

### Prerequisites

- Git (latest stable)
- Node.js 20.x or newer 
- npm 10.x or newer

Verify your environment:

```bash
node -v
npm -v
git --version
```

### Add-ons

- `react-router-dom`: Client-side routing between Home, In, Out, and Mixed pages.
- `recharts`: Data visualization components (pie charts, bar charts, tooltips, responsive containers).
- `tailwindcss`: Utility-first styling.
- `postcss`: CSS transformation pipeline used by Tailwind.
- `autoprefixer`: Adds vendor prefixes for browser compatibility.
- `react-scripts`: Build/dev/test scripts for the Create React App toolchain.

### Installation Steps

Copy/paste the commands below to install and run the project from a fresh clone:

```bash
git clone https://github.com/WSU-CPTS322-SP26/Visual-Network.git
cd Visual-Network
make setup
make build
make run
```

#### Data setup (real vs seed/mock)

- Current state: no backend/database seed step is required.
- Chart values are mock data defined directly in:
	- `src/InPage.js`
	- `src/OutPage.js`
	- `src/MixedPage.js`
- To use real data, replace the local `data`/`barData` arrays in those files with API-fed or packet-capture-fed values.

## Functionality

1. Start the app with `npm start`.
2. Open the app in your browser (default: `http://localhost:3000`).
3. Use the top navigation:
	 - **Home**: Overview of the project and displayed metrics.
	 - **In Page**: Inbound traffic focus (Volume In, Protocol, Latency).
	 - **Out Page**: Outbound traffic focus (Volume Out, Protocol, Latency).
	 - **Mixed Page**: Combined in/out perspective with protocol and latency charts.
4. Hover chart segments/bars to inspect values.

## Known Problems

- **Mock data only**: The app currently visualizes hardcoded sample values, not live networking values.
	- Location: `src/InPage.js`, `src/OutPage.js`, `src/MixedPage.js`
- **Hard-coded router basename**: Navigation uses a fixed basename (`/WSU-CPTS322-SP26/Visual-Network`) that may require adjustment for local/custom deployments.
	- Location: `src/App.js`

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :D

## Additional Documentation

Currently, no additional project docs are committed.

Suggested docs to add under a future `docs/` directory:

- Sprint reports
- Architecture notes
- User guide / demo links

## License

This repository should include a `LICENSE.txt` file.

Recommended license: MIT

- https://choosealicense.com/licenses/mit/

