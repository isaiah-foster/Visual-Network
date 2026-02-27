# Sprint 1 Report (2/7/26 - 2/24/2026)

## What's New (User Facing)

### Admin Features
* Feature 1: Easy setup with setup.sh
* Feature 2: Create login credentials for users

### General User Features
 * Feature 1: Login page
 * Feature 2: Ability to login with admin provided credentials
 * Feature 3: Network metric viewing page


## Work Summary (Developer Facing)
We finalized the tech stack required to implement our idea into code. Application modules were assigned to team members based on our previous experience and our fields of interest. We implemented a Go backend to support the following list of capabilities:
  1. Collect ingress/egress network data from the Linux kernel network stack.
  2. Aggregate network data into cohesive JSON.
  3. Establish connection to frontend and forward network data.
  4. Receive login requests from frontend, query/update the database, and respond.
We implemented a frontend to support the following:
  1. User logins
  2. Visual display of all network metrics seperately for ingress and egress
We implemented an SQL database to store user login information as well as an Nginx service to handle http requests and establish websockets connections from frontend to backend. We had an extremely steep, but productive learning curve to overcome with understanding eBPF and how to implement it. We also had to learn how to build an auth system from scratch, which took a nontrivial amount of research.

## Unfinished Work
If applicable, explain the work you did not finish in this sprint. For issues/user stories in the current sprint that have not been closed, (a) any progress toward completion of the issues has been clearly tracked (by checking the checkboxes of  acceptance criteria), (b) a comment has been added to the issue to explain why the issue could not be completed (e.g., "we ran out of time" or "we did not anticipate it would be so much work"), and (c) the issue is added to a subsequent sprint, so that it can be addressed later.

Since this is the first sprint, there are many ideas and features that are still in the oven that'll be ready for the next sprint. We have Issues that we made that might not be finished by the sprint deadline. Some of these include:

Issue 1.) Improving Frontend UX and Visuals. (Didn't have enough time to sharpen up the visuals for this sprint deadline due to other bigger priorities)


## Completed Issues/User Stories
Here are links to the issues that we completed in this sprint:

 * https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/9
 * https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/1


 Reminders (Remove this section when you save the file):
  * Each issue should be assigned to a milestone
  * Each completed issue should be assigned to a pull request
  * Each completed pull request should include a link to a "Before and After" video
  * All team members who contributed to the issue should be assigned to it on GitHub
  * Each issue should be assigned story points using a label
  * Story points contribution of each team member should be indicated in a comment
 
 ## Incomplete Issues/User Stories
 Here are links to issues we worked on but did not complete in this sprint:
 
 * https://github.com/WSU-CPTS322-SP26/Visual-Network/issues/6 (Decided that this feature was not worth the time for this sprint as its only a visual enhancement and we had other priorities such as getting a connecting the frontend and backend)
 * URL of issue 2 <<One sentence explanation of why issue was not completed>>
 * URL of issue n <<One sentence explanation of why issue was not completed>>
 
 Examples of explanations (Remove this section when you save the file):
  * "We ran into a complication we did not anticipate (explain briefly)." 
  * "We decided that the feature did not add sufficient value for us to work on it in this sprint (explain briefly)."
  * "We could not reproduce the bug" (explain briefly).
  * "We did not get to this issue because..." (explain briefly)

## Code Files for Review (top 5-6 files that are highlight / best files)
Please review the following code files, which were actively developed during this sprint, for quality:
 * InPage.js (https://github.com/WSU-CPTS322-SP26/Visual-Network/blob/lucas-branch/src/InPage.js)
 * [Name of code file 2](https://github.com/your_repo/file_extension)
 * [Name of code file 3](https://github.com/your_repo/file_extension)
 
## Retrospective Summary
Here's what went well:
  * Team communitcation and steady progress throughout the duration of the sprint.
  * Completed issues that needed to be done before sprint 1.
  * We have plans for the future on what to improve for the next sprint.
 
Here's what we'd like to improve:
   * Visuals and UX on the frontend side of things.
   * Item 2
   * Item x
  
Here are changes we plan to implement in the next sprint:
   * Item 1
   * Item 2
   * Item x
