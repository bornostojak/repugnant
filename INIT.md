## RULES OF ENGAGEMENT
I have an idea for a new project, but you have to help me flesh it out. I can only give you the outline of what I want and I require of you to do the fillowing:
- flesh out the raw idea
- ask me detailed questions one by one, so as to not overwhelm me with 20 questions at once about how I'd like to have the program work, structure the code, what tool you'd recommned to use and why i should allow you to use the (why they are the tools for the job i.e. regular expression matcing etc.)
- treat this project as a monorepo, since other tool-related code may be added in the fulre (ie. code for visual studio plugin)
- git - just use it where it makes sense, use branches for writing featrues and merge them into the mian branch with `git merge --no-ff`
- only once the prep has been completed wait for my signal to start the work and implementation of code
- consider where each feature runs (frontend/backend) and where the infrastrucutre is (is there need for caching with reddis
    * don't overdo it - you could thenically add caching to everything but if the app will not have more than 100 users at once, if it's not strictly specified or if the feature doesn't strctly require it, then you can skipp reddis, or replace postgresql with sqlite3 but leave room for easy impelemtation (keeping this in mind when chossing to use rules like the **registry pattern**)
- make sure to incorpore the follwing
    - detailed logging with log-level selection (during testing, use the least-output intensive one until debugging requires exageration)
    - make the logs sensible for helping in finding bugs duriong production, but not have every log be a treace log - make them make sense
    - frontend interfaces should be responsive for mobile as well
    - browser-based testing of frontends - use chromium or such to test how the page behaves
    - use standardized config syntax
        * .env files and environment variables for configuring backend behavior - consider it will probably run in docker

### PREP - Markdown files
Add the follwing markdown files:
- SPEC.md - flesh out a detailed specification once you're ready - the full specifications for how the project will work, what the tool must do, what it musn't ever do, how it should behave and a full fledge deterministic ruleset for the tool.
- TASTE.md - specify designe choices, params about styling the frontend, backend code strcutre and behavior like logging preferences, and very very detailed stuff that should be referenced in the AGENTS.md to always remember to make sure the code will allign with the wishes in the TASTE.md - this should help keep the code consistent in all manner of applicatios and help prevent stuff like having a link for someting on page A and the same something on page B be a button, stuff like using the **registry-pattern** would also go here, etc.
- AGENTS.md - and specify how each of the .md files will be treated
- VISION.md - a simple explaination of the goal and vision for the project to keep the coding focused on the result, not just solving a coding issue, explain what issue the code will solve, what it is usefull for and what the people that will use it expect it to do and what it will do for them
- STYLE.md - for frontend applications, write consistent rules that need to be adhered to (like css guidelines for card, should the app use cards-style interface or avoid them, etc.)
- DOCS.md - instructions on how to writie doumentation for the code - after each features has been implemented, document it in detaild, how it's being used, how it's being set-up or configured, how to maintain it and the logic behind the featrue, what the user/customer will gain from the feature
- INFRASTRUCTURE.md - will it be running in docker, how the different modules connect, wheich module takse responsibility for which pare of the app's logic - all of that and more goes here
- README.md - place this in the project root and explain in detail what the project does, the project's aspirations and goals, how to install it, how to siply configure it (and where to look for detailed config instructions), where the detailed docs are
- make sure to ephasize that design chocises and their changes/updates must be updated or added to ALL RELEVANT .md FILES
- prepare detailed TESTS.md - think of as many possible points of failure as you can, that would make sense to test for, but they have to make strategic sense, so not just 2-3, 20 or more is what I'd consider depending on the complexity of the project, but keep in minda that if 12 cover 99% of the behavior, comming up with the 8 others neede to reach 20 is just a waste of computation
- place the AGENTS.md in the root of the project and create `rules_of_engagement` dir to keep te other .md files organised there - of course all .md files must be mentioned, along with their purpose in the AGENtS.md file
    * if you think it splitting the `r_o_e` into `r_o_e/frontend`, `r_o_e/backend`, `r_o_e/edge` would make the LLM context less congested (why do I need to fill my context with frontend rules for when I'm working on the backend?) then do it
    * make sure that consistency is being maintained across all rules and instructions in these markdown files

## EXECUTION and TESTING
- you are hereby given permission to execute the code you are writing on this machine in order to test and validate that your implementation works
- run the code and observer how it behaves via logs and performance-wise where possible
- use the browser to inspect the frontend look and feel - navigate all pages and makins sure they've been implemented in a propper way, look for stylistic inconsistencies

## SELF-IMPROVEMENT
- once you've completed the task, asses what you've done
- think of ways that the project may be improved - feature-wise, performance-wise, look-wise, design-wise...
- create feature branches for the improvements you think and feel free to spawn new branches to try them out
- create a `feature-merge` branch once you've confident that the improvement features you've come up with are ready to be presented
- DON'T MERGE THEM INTO main BRANCH UNTIL I'VE GOVE OVER THEM AND GIVEN YOU CREDIT FOR WHAT YOU'VE DONE - the feature may be excelent or it could be pivoted to another use

## CONTINUOUS IMPLEMENTATION
- THIS MUST BE DONE AS I INSTRUCT YOU
    * I'll be leaving you unattended once the coding starts
    * you are to operate autonomously from then on untill you've finished ALL OF YOUR TASKS
    * you are not to stop after each git commit - do you building and testing after each feature, commit and merge when the feature is ready, then move on to the next feature
    * if you are higgin the 5h token limit, if you can - temporarily stop the exectuion of your task until the 5h limit reset, or do something like a backround script to restart the exectuion on 5h roll-over
    * validate all features by testing and observing them, do your `go test` and all, but do actually USE the feature to see of it works "in production"

## TASK
I'm calling this project rePugnant (or just repugnant - the P is styliesic). 
The goal: simplify the writing of good documenation, that is kept in one place like a /docs dir or a web+mobile app, during the code writing process.
Programmers love to write code, including comments, because it's easy to write once you're in the zone.
Think of writing code in neovim/vim - you know the shortcuts and you are fast, but using stuff like confluce, allthough it is pretty, is very slow, and pulls you out of the coding flow state.
The goal of this project is to change that
Here is a very basic outline of how it'd work:

### HOW IT WORKS
+ take code comments and documentation for generally complex code
+ It's "easy" writing comments
+ It's hard writing good docs for code
+ We take a syntax like

```c
// $rPg: handling reddis cache resolution
// $~ some
// $~ markdown
// $~ here

code {
    ...

    // ?rPg: very tricky syntax
    // $~ with detailed md explain here
    tricky_code_to_be_quoted {
    }
    // !rPg_end_quote

};
```

+ You add a git pre_commit hook to parse over the entire code and look for those
+ rePugnant takes the snippets, generates a signed documentation file in /docs and/or on a web server, assigns it a feature code, or the user does like $rPg(redis, here_add_search_tags) and then replace this with

```c
// docs: #ROOT/docs/....
or
// docs: https://rpg.companysite.com/d/project/feature

code {
    ...

    // docs: #/feat:line:line
    or
    // docs: https://rpg.companysite.com/d/project/feature#heading
    tricky_code_to_be_quoted {
    }
    // !rPg_end_quote

};
```

+ You could also have a rpg.comf file, define rules there, and define stuff like docs root, domain, etc. so you could remove the #ROOT
+ Later also add tracking of changes - some code may be changed so if the parser realizes this, it can request a change comment to be added

### TOOLS

- backend: use go
- frontend: 
    * react or svelte+runes
    * typescript
    * tailwind
    * don't use cards
    * make the designe minimalistic and modern
    * add a touch of color (don't ovrdo it with color)
    * responsive
- cli: go

### TASKS
- improve upon this concept
- for now keep this simpleistic, but if you see issues I've overlooked warn me
- write the CLI tool for parsing this, including the `rpg --commit-hook` to use it as the hook 
- for stuff that I'm not providing enoght clarification - ask now during the initiation stage


