# Create a zoom app

Visit the [zoom marketplace](https://marketplace.zoom.us/).

## Create zoom application. 


![App_Marketplace](https://user-images.githubusercontent.com/6961/222057177-c388df1b-4b49-4555-8867-535c86affe13.png)


## Choose Webhook Only Application

![Choose Webhook Only](https://user-images.githubusercontent.com/6961/222057317-d87498ff-13eb-4b77-b1ff-2808e52e9e2a.png)

## Give your application a name

![Give it a name](https://user-images.githubusercontent.com/6961/222057676-2af00cab-416f-49f2-8430-78ae8d7f5512.png)

## Fill out the application information. 

![Fill out](https://user-images.githubusercontent.com/6961/222057800-7c77061e-6f9f-47f4-aba8-07d5d2980aa9.png)

## Click continue. 


![Zoom app](https://user-images.githubusercontent.com/6961/222058197-a71f9aad-c673-470b-a8e5-3f642c64ed6a.png)

## Toggle Event Subscriptions

![Toggle Event Subscriptions](https://user-images.githubusercontent.com/6961/222058339-63fdf6ee-b00c-4b86-ac12-34af57609ee8.png)


## Press + Add Event Subscriptions



You'll need a public endpoint for a listener -- which is where you'll run this app. You'll see a "validate" button. Zoom send validation payload through every so often for authentication and you need to respond to them with. The app has this built in and you just need to provide your `ZOOM_SECRET` (which is `secret token` on the zoom app configuration page).  

![App_Marketplace](https://user-images.githubusercontent.com/6961/222058586-1fdf418a-4255-4325-bb0e-d9adcfc12c2c.png)

## Press the "Add Events" button.

![zoom_event_types](https://user-images.githubusercontent.com/6961/222058647-4c058c29-5573-462a-a7f0-607c9e9a26b6.png)


Click on validate. If the webhook can't be validated, see the configuration portion of the documentation. 

## Save.
