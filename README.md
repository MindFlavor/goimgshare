[![Build Status](https://travis-ci.org/MindFlavor/goimgshare.svg?branch=master)](https://travis-ci.org/MindFlavor/goimgshare) [![stable](http://badges.github.io/stability-badges/dist/stable.svg)](http://github.com/badges/stability-badges) [![Coverage Status](https://coveralls.io/repos/github/MindFlavor/goimgshare/badge.svg?branch=master)](https://coveralls.io/github/MindFlavor/goimgshare?branch=master)

# GO Image and video share

GO Image and video share is a program that creates a file sharing website useful to share images and other files. The main features are:
- Support for Facebook authentication
- Support for Google authentication
- Support for GitHub authentication. 
- Automatic image resize for faster thumbnails (with cache)

## How install

#### Check GO installation
Ensure to have GO installed (tested with 1.5). If not sure issue: 
```
go version
``` 
from command line. Make sure to have a valid ```$GOPATH``` environment variable. 
#### Get the source code
```
go get -u github.com/mindflavor/goimgshare
```
This will get the source code. You can use the same command later to update the source code.

#### Copy sample configuration files
This step is optional, you can edit the original files without copying them. In that case, however, you will have to stash your changes before updating the source code.
``` 
cp $GOPATH/src/github.com/mindflavor/goimgshare/config.json /etc/goimgshare/config.json
cp $GOPATH/src/github.com/mindflavor/goimgshare/shared_folders.json /etc/goimgshare/shared_folders.json
```
Of course replace ```/etc/goimgshare``` with your path. Windows users should use a Windows path of your choosing.

#### Edit the configuration files
There are two configuration files, one for the general configuration and another for the shared folders. 
##### Main configuration file
```json
{
  "Port":8080,
  "CacheInternalHTTPFiles":false,
  "LogInternalHTTPFilesAccess":true,
  "InternalHTTPFilesPath": "D:\\GIT\\Go\\src\\github.com\\mindflavor\\goimgshare",
  "SharedFoldersConfigurationFile":"D:\\temp\\shared_folders.json",
  "ThumbnailCacheFolder":"D:\\temp\\thumbs",
  "SmallThumbnailSize":500,
  "AverageThumbnailSize":1000,
  "Google":{
    "ClientID":"247709547460-aovc4l7ghgs2senq0cr2ivphiu12kssv.apps.googleusercontent.com",
    "Secret":"h7tgEmIMHf6Q-xvQQcP9glmQ",
    "ReturnURL":"http://localhost:8080/auth/google/callback"},
  "Facebook":null,
  "Github":null
}
```
The attributes are:

Attribute | Explanation
 ---- | ----
 ```Port``` | The webserver port. If the port is already in use the program will panic. 
 ```CacheInternalHTTPFiles```  |If ```true``` the webserver will cache in memory the static HTML an JS files. ```false``` is useful if you are editing the files because you don't need to restart the program to see the changes. 
```LogInternalHTTPFilesAccess``` | If ```true``` the console will log the access to static HTML and JS files.
```InternalHTTPFilesPath``` | This folder must point to the ```goimgshare``` root folder. It will be ```$GOPATH/src/github.com/mindflavor/goimgshare``` unless moved.
```SharedFoldersConfigurationFile```|The shared folders configuration file. It will be read at startup so if you change it you will need to restart the program.
```ThumbnailCacheFolder```|Thumbnail cache folder. It must be a valid path.
```SmallThumbnailSize```|Thumbnail size in pixels. Smaller images are faster but of course will be grainy on high-resolution displays. Note that you need to clean the ```ThumbnailCacheFolder``` manually if you change this.
```AverageThumbnailSize```|Unused at the moment.
```Google::ClientID```|Your Google client ID. 
```Google::Secret```|Your Google client secret. 
```Google::ReturnURL```|Google authentication return URL. Do not change the suffix ```/auth/google/callback```.
```Facebook::ClientID```|Your Facebook client ID. 
```Facebook::Secret```|Your Facebook client secret. 
```Facebook::ReturnURL```|Facebook authentication return URL. Do not change the suffix ```/auth/google/callback```.
```Github::ClientID```|Your Github client ID. 
```Github::Secret```|Your Github client secret. 
```Github::ReturnURL```|Github authentication return URL. Do not change the suffix ```/auth/google/callback```.

As you can see, the authentication providers are optional. Just make sure to specify at least one or you won't be able to log in! :wink:.

##### Shared folders configuration file
This is a sample shared folders configuration file:
```json
[
  {
    "ID": "001",
    "Name": "temp on D",
    "Path": "D:\\temp on D",
    "AuthorizedMails": {
      "francesco.cogno@gmail.com": true,
      "valentina.test@gmail.com": true
    }
  },
  {
    "ID": "002",
    "Name": "temp on C",
    "Path": "C:\\temp",
    "AuthorizedMails": {
      "francesco.cogno@gmail.com": true,
      "prova@test.com": true
    }
  }
]

```
The file is an array of shared folders. The first folder can be viewed - along with its contenints - by      francesco.cogno@gmail.com and valentina.test@gmail.com. The second folder is also accessible by francesco.cogno@gmail.com but not by valentina.test@gmail.com. It's the only folder accessible by prova@test.com.

The shared folder fields are:

Attribute | Explanation
 ---- | ----
ID | An arbitrary ID. It can be anything you want as long as it is unique in the array.
Name | The user-facing folder name. It does not have to be unique (but it will be confusing if it is not).
 Path | The path of the folder. The folder should be accessible and the files within should be readable by the program's account.
AuthorizedMails | Is a map of emails and a ```boolean``` indicating if the mail has access. Any email not specified here is not allowed to access the folder (and will not see it either in the list). The same happens if you specify false as mail attribute (useful if you want to temporarily revoke user access).

#### Build and install the program
```
go install github.com/mindflavor/goimgshare
```
This will place the binary in your ```$GOPATH\bin``` folder.
#### Start the program specifying the main configuration file
```
$GOPATH/bin/goimgshare /etc/goimgshare/config.json
```

If everything goes as planned (ie there are no bugs and the configuration files are correct) you should see something like this:

![](http://i.imgur.com/YSssxWY.jpg)

And you will be able to browse the http://localhost:<port> address. Since you are not authenticated you will see the authentication page (the number of authentication providers is dependent on the configuration - check the [authentication_providers] section for details):

![](http://i.imgur.com/E2nKiGH.png)

Here we have Google and Facebook as authentication providers. Note that the links are dynamic and dependent on the configuration (so you won't see a button if you have not configured the relevant provider). 

### Acknowledgements 
This program uses these libraries:
* [https://github.com/stretchr/gomniauth](https://github.com/stretchr/gomniauth)
* [https://github.com/stretchr/signature](https://github.com/stretchr/signature)
* [https://github.com/gorilla/mux](https://github.com/gorilla/mux)
* [https://github.com/nfnt/resize](https://github.com/nfnt/resize)
* [AngularJS](https://angularjs.org/)
* [Bootstrap](http://getbootstrap.com/)


----------

**Happy coding**

[Francesco Cogno](francesco.cogno@outlook.com)
