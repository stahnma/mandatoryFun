# Zoom API Enablement

To enable to the Zoom notifier including the link to the meeting including the passcode, you'll need a second application.

## Go to the Zoom Marketplace

Visit the [zoom marketplace](https://marketplace.zoom.us/).

Log in with your credentials.

## Create zoom application

![App_Marketplace](https://user-images.githubusercontent.com/6961/222057177-c388df1b-4b49-4555-8867-535c86affe13.png)


Choose Server to Server OAuth application.


![App_Marketplace-2](https://github.com/user-attachments/assets/dc8fef45-dc6a-4ab1-8c4e-d47e8add0011)


Give your application a name, something like "lookup meeting info" would be ideal, so you remember later what it does.


## Configuration the application

![App_Marketplace-3](https://github.com/user-attachments/assets/6746e3ad-dd82-4bc4-8725-aca696c31e4b)


Once the application is created, you can see the Account ID, Client ID, and Client Secret. You will need these in your configuration for zoom notifier.

:camera: Record the required environment variables.

From there, you need to fill out basic information for the application.

![App_Marketplace-4](https://github.com/user-attachments/assets/143f2c51-0462-48bb-b797-1529c4ee4ed3)


Once you fill out that information, you'll get to the screen with the Tokens. Secret Token is what you need.

# Scope the Application

![App_Marketplace-6](https://github.com/user-attachments/assets/2d15490d-39bc-4b16-9ff4-aead45e3c2d3)


From there, you'll need to set scopes. The screenshot is probably the easiest way to see what you need. 

- [ ] dashboard:read:list_meeting_participants:admin
- [ ] dashboard:read:list_meetings:admin
- [ ] dashboard:read:list_meeting_participants:master
- [ ] meeting:read:meeting:admin
- [ ] meeting:read:invitation:admin
- [ ] meeting:read:list_meetings:admin
- [ ] report:read:meeting:admin
- [ ] report:read:list_meeting_participants:admin

:warning: It is possible this is more scope than needed. Some items may have been added during development and not pulled back yet. Feel free to adjust the scopes and if you do, please open a PR to update this document.

Set scopes, and then "Activate" the application.

# Setup your configuration file

From there, you'll need to add the environment variables documented in the primary [README.md](../README.md) file.
