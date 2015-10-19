# GO Image and video share


GO Image and video share is a program that creates a file sharing website useful to share images and other files. The main features are:
- Support for Facebook authentication
- Support for Google authentication
- Support for GitHub authentication. 
- Automatic image resize for faster thumbnails (with cache)

## How install

1. Ensure to have GO installed (tested with 1.5). If not sure issue 
```
go version
``` 
from command line. Make sure to have a valid ```$GOPATH``` environment variable. 
2. Get the source code:
```
go get -u github.com/mindflavor/goimgshare
```
This will get the source code.
3. Copy the sample configuration files where you want them to be. This step is optional, you can edit the original files without copying them. In that case, however, you will have to stash your changes before updating the source code.
``` 
cp $GOPATH/src/github.com/mindflavor/goimgshare/config.json /etc/goimgshare/config.json
cp $GOPATH/src/github.com/mindflavor/goimgshare/shared_folders.json /etc/goimgshare/shared_folders.json
```
Of course replace /etc/goimgshare with your path. Windows users should use a Windows path of your choosing.
4. Edit the configuration files. On how to do that consult the (config_files) section.
5. Build and install the program:
```
go install github.com/mindflavor/goimgshare
```
This will place the binary in your ```$GOPATH\bin``` folder.
6. Start it specifying the main configuration file:
```
$GOPATH/bin/goimgshare /etc/goimgshare/config.json
```

If everything goes as planned (ie there are no bugs and the configuration files are correct) you should see something like this:

And you will be able to browse the http://localhost:<port> address. Since you are not authenticated you will see the authentication page (the number of authentication providers is dependent on the configuration - check the (authentication_providers) section for details):

