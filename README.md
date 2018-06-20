# telegram-deploy-bot

telegram-deploy-bot
===================

Example:

    export BOT_JENKINS_URL=http://jenkinshost:8080
    export BOT_JENKINS_USER=admin
    export BOT_JENKINS_PASS=put_password_here
    export BOT_JENKINS_INPUT_ID=put_pipeline_input_id_here
    export BOT_DEBUG=true
    export BOT_TOKEN=put_telegram_bot_token_here
    
    telegram-deploy-bot

telegram-bot-send
=================

Example:

    export BOT_TOKEN=put_telegram_bot_token_here

    telegram-bot-send jenkins_job_name jenkins_build_id telegram_chat_id
