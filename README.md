# Smokestopper

A simple web application designed to help anyone stop smoking
The application is structured as follows and open source:

- Svelte frontend
- Go backend
- Writes to a badger Database
- Uses jwt to active accounts

The idea is simple, you enter your email address and click a url to verify that it is your email address.
As soon as it is verified, the time is saved in the database and it defines the moment you stopped smoking.
Periodically you will receive updates about how long it has been since you stopped smoking and what the benefits on your health are.

I was also thinking of creating a leaderboard of some sort, and defining a moment where, according to the application, you can be defined a non smoker.
Furthermore, I would like to implement forum based interaction between users, for when one would like to share a story, or a word of advice to help others to stop smoking.
