<h1 align="center">J.T. Slackbot</h1>

<div align="center">
  A small Go app that integrates with <a href="https://slack.com/">Slack</a> to power a virtual J.T. to hang out in your workspace.
</div>

<br />

<div align="center">
  <img src="https://img.shields.io/github/go-mod/go-version/drewnorman/jt-slackbot" />
  <img src="https://img.shields.io/github/license/drewnorman/jt-slackbot" />
</div>

<br />

<div align="center">
  <sub>
    Brought to you by <a href="https://foxfuelcreative.com/team/drew-norman">Drew.0</a> with inspiration from the lovable <a href="https://www.instagram.com/p/Bp2PfjSBPXr/">J.T.</a>
  </sub>
</div>


### Features
- __Intelligent Responses:__ Mention the good boy himself in a message with `@J.T.` and he will try to respond appropriately. No guarantees. After all, he is only a pup.
- __Real-Time Interactions:__ With a real-time connection to Slack via <a href="https://api.slack.com/apis/connections/socket">Socket Mode</a>, it's like J.T. is really talking to you! Omg!
- __Public Channel Infiltration:__ On start up, J.T. Slackbot will try to join all of your public channels. He really just wants some company.
- __Automatic Connection Renewal:__ J.T. is infinite! J.T. is eternal!


### Installation

Clone the project:

    git clone https://github.com/drewnorman/jt-slackbot.git

From the project root, create a copy of the configuration template:

    cp .env.template .env

Then replace the values for `SLACK_APP_TOKEN` and `SLACK_BOT_TOKEN` with your corresponding Slack application tokens.


### Usage
Start the application.

    go run .

Or with Docker Compose:

    docker-compose up -d