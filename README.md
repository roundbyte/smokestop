# Smokestopper

A simple web application designed to help anyone, including me, in the quest to stop smoking. At the same time I consider this an exercise for some web server programming.

## Step 1 - User Registration (done)

-   Registration form:
    -   Email Address
    -   Username
    -   Password
-   It will be required to verify the email address.
-   After verification, login can proceed, and makes use of a login cookie.

## Step 2 - Design the application (in progress)
A big button in the middle of the screen that saved the current timestamp and defines the moment you stopped smoking.
Periodically you will receive updates about how long it has been since you stopped smoking and and especially what the benefits on your health are.

## Step 3 - Brainstorm
- Add a leaderboard of some sort.
- Add forum based interaction between users, for when one would like to share a story, or a word of advice to help others stop smoking.

You might find this soon at jakobthesheep.com

## Technology

- Backend written in Go
- BadgerDB for the database choice
- Svelte will be used for frontend

You must create a .env file should you want to test this, containing:
SERVERPORT= on which port to listen
DBPATH= path to a folder of choice which will become the DB
SECRETKEY=for the cookie store
MAILPSW=needed by the mailer to relay mail through gmail