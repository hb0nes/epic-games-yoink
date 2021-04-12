# What is this?
This repository can serve as a example for your own code, or can be used as a full solution.  
The program will log into your account on epicgames.com and add all free games that it finds to your account.  
No data is shared with third parties or me.  
# Getting Started
There are two methods of automatically adding all of these games, manual and automatic.
## Automatic
- [Fork](https://github.com/hb0nes/epic-store-free-games-snatcher/fork)  this repository
- Add three or more GitHub secrets:
  - hCaptchaURL
  - username
  - password
  - OTPSecret
  - telegramID - You can get your telegramID by sending a message to @EpicGamesYoinkBot
  - IMGUR_CLIENT_ID - In case you want to see screenshot URLs in the logs, you can insert your own imgur api credentials, but it isn't necessary
  - IMGUR_SECRET
  - IMGUR_REFRESH_TOKEN
- Enable [Git Actions](https://docs.github.com/en/free-pro-team@latest/actions/managing-workflow-runs/disabling-and-enabling-a-workflow) and enable the 'Go' workflow.
- It will run twice per day, each day. Do not adjust the interval, hCaptcha has very strict ratelimits.  
## Manual - I would seriously just go for Automatic...
- Fetch the proper [release](https://github.com/hb0nes/epic-store-free-games-snatcher/releases). 
- Go [here](https://dashboard.hcaptcha.com/signup?type=accessibility) and get an accessibility URL by filling in your e-mail. The mail that they send you contains a button.  
Right click this button and get the URL, and insert that URL in config.yaml.example    
This example file also shows what it should look like. If you want to use one key, remove the example so that you only have one entry left (your own key).  
- Fill in your email/username & password as well, to be able to log into the Epic Games store.  
- When using 2FA, fill in the secret that is given during the 2FA setup or find some other way to retrieve the secret.
- Rename config.yaml.example to config.yaml and make sure it is in the same directory as your executable.
- Make sure you have Chrome installed, and run the program!  

If it all fails, open an issue and I might look into it.
I might also not.
