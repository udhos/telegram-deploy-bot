# telegram-deploy-bot

telegram-deploy-bot
===================

Service controlling the bot.

Example:

    export BOT_JENKINS_URL=http://jenkinshost:8080
    export BOT_JENKINS_USER=admin
    export BOT_JENKINS_PASS=put_password_here
    export BOT_DEBUG=true
    export BOT_TOKEN=put_telegram_bot_token_here
    export BOT_AUTHORIZED_USER_ID_LIST=userId1,userId2,...,userIdN
    
    telegram-deploy-bot

telegram-bot-send
=================

Send a message thru the bot.

Example:

    export BOT_TOKEN=put_telegram_bot_token_here

    telegram-bot-send jenkins_job_name jenkins_build_id jenkins_input_id telegram_chat_id
