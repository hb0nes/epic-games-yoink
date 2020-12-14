# What is this?
This repository can serve as a example for your own code, or can be used as a full solution.  
The program will log into your account on epicgames.com and add all free games that it finds to your account.  
It uses hCaptcha's own hCaptcha bypass mechanism B-).  
No data is shared with third parties or me.
# Getting Started
There are two methods of automatically adding all of these games, manual and automatic.
## Manual
- Clone the repository
- Go to https://dashboard.hcaptcha.com/signup?type=accessibility and get an accessibility URL by filling in your e-mail. The mail contains a button.  
Right click this button and get the URL to insert it into config.yaml.  
The example file shows what it should look like.  
- Fill in your username & password as well to use for logging into Epic Games store.  
- Rename config.yaml.example to config.yaml
- Make sure you have Chrome installed, and run the program.  
## Automatic
- Fork this repository
- Add three GitHub secrets:
  - hCaptchaURL
  - username
  - password
- Unstar/Star your own repository. A GitHub action is now started that will take care of everything for you.  
If it fails, open an issue and I might look into it.
