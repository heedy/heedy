# Android App

<a href='https://play.google.com/store/apps/details?id=com.connectordb_android&pcampaignid=MKT-Other-global-all-co-prtnr-py-PartBadge-Mar2515-1'><img width="250" alt='Get it on Google Play' src='https://play.google.com/intl/en_us/badges/images/generic/en_badge_web_generic.png'/></a>

<h4 style="color: red;padding-bottom: 40px;">This document refers to a pre-release version of the app (To be released soon).<br/> You can try this new version <a href="https://play.google.com/apps/testing/com.connectordb_android">by joining the beta.</a></h4>

When setting up the Android app, you will need to be on a network with access to your ConnectorDB server. That is, you should be connected to your home wifi if running ConnectorDB at home.

After installing the app, you will be presented with this screen:

<img src="/assets/docs/img/android-login.png" width="300"/>

### Finding Server URL

If running ConnectorDB locally, you need to find the IP address at which your server is running. This can be done from the ConnectorDB interface
by going into the top right menu, and selecting "Server Info".

<img src="/assets/docs/img/top-menu.png"/>
<img src="/assets/docs/img/server-info.png"/>

The Server field is what you should put as the server in the android app. If running a private server on your network, you should also enable the option to only sync when connected to your wifi network:

<img src="/assets/docs/img/android-login-filled.png" width="300"/>

### After Login

After logging in, you should go to the options tab, and enable/disable whichever loggers you would like. Please note that certain loggers, such as logging steps and other data from google fit, or logging sleep from Sleep As Android will source their data from proprietary sources, meaning that Google or other third parties might have access to this data!

If you don't want this, disable loggers from google fit!
