# Sprint  Report (2/25/26 - 4/4/2026)

## What's New (User Facing)

### Admin Features
* Feature 1: Secure the platform with login setup
* Feature 2: Easy setup with setup.sh
* Feature 3: Create login credentials for users

### General User Features
 * Feature 1: Login page
 * Feature 2: The frontend now displays live network data from the host machine.
 * Feature 3: The user may now view pie charts to visualize different network protocol usage over ingress and egress traffic.
 * Feature 4: The user may now view pie charts to visualize the packet volume of top n ip addresses.
 * Feature 5: The user may now view bar graphs to show network latency to top n most used ip addresses.
 * Feature 6: Users may logout of account manually

## Work Summary (Developer Facing)
We completed support for the frontend to display live data through our reverse proxy from the backend. This data replaces the previous dummy/mock data we had. We also m
  1. Collect ingress/egress network data from the Linux kernel network stack.
  2. Aggregate network data into cohesive JSON.
  3. Establish connection to frontend and forward network data.
  4. Receive login requests from frontend, query/update the database, and respond.
We implemented a frontend to support the following:
  1. User logins
  2. Visual display of all network metrics seperately for ingress and egress
We implemented an SQL database to store user login information as well as an Nginx service to handle http requests and establish websockets connections from frontend to backend. We had an extremely steep, but productive learning curve to overcome with understanding eBPF and how to implement it. We also had to learn how to build an auth system from scratch, which took a nontrivial amount of research.

## Unfinished Work

Issue 1.) Improving Frontend UX and Visuals. While the code is written for this part, and we are happy with the new look, we did not have time to merge it into the other commits prior to the sprint deadline. This will be completed early in sprint 3.

## Completed Issues/User Stories
Here are links to the issues that we completed in this sprint:

 * https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/10
 * https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/3
 * https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/12

 
 ## Incomplete Issues/User Stories
 Here are links to issues we worked on but did not complete in this sprint:
 
 * https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/6 (Require merging with main)
 
## Code Files for Review (top 5-6 files that are highlight / best files)
Please review the following code files, which were actively developed during this sprint, for quality:
 * (InPage.js) (https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/lucas-branch/src/InPage.js)
 * (tcm_monitor.c) (https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/backend/traffic-aggregation/ebpf/tc_monitor.c)
 * (mysql_store.go) (https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/main/backend/auth_service/store/mysql_store.go)
 
## Retrospective Summary
Here's what went well:
  * Team communitcation and steady progress throughout the duration of the sprint.
  * Completed most of the issues that needed to be done before sprint 2.
  * We have plans for the future on what to improve for the next sprint.
 
Here's what we'd like to improve:
   * Better formatted layout of metric displays
   * Improve the speed it takes the frontend to update with new data
   * Keep code more modular for easier merging
  
Here are changes we plan to implement in the next sprint:
   * Separate the frontend from the backend, so the backend daemon can be more light weight, with the frontend being run on a monitoring machine.
   * Verify that no bugs exist during sprint 3
   * Add a ping option from the frontend to test latency to different servers
   * Streamline the user experience of the app
   * Allow admin users to edit user account information and remove users
   * Possible implement forced admin password update after first login.