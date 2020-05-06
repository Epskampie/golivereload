`golivereload` is a command line tool for webdevelopment that automatically refreshes your browser whenever you change a file.

Quickstart
----------

* In your terminal, go to the directory you want to watch for changes.
* Run the `golivereload` command.
* Include this script in your html: 
`<script src="http://0.0.0.0:35729/livereload.js"></script>` 
or use the [livereload extension](https://chrome.google.com/webstore/detail/livereload/jnihajbhpnppcggbcgedagnkighmdlei?hl=en) for your browser.
* Done! The page refreshes whenever you change a file.

Installation
------------

* Download the [newest release](https://github.com/Epskampie/golivereload/releases) for your OS.
* Install it per project:
    * Unzip the binary to your project dir.
* Or install it globally:
    * linux: copy the binary to the `/usr/local/bin` directory.
    * windows: 
        * Create a new `bin` directory in your home folder, then [add it to the PATH](https://docs.alfresco.com/4.2/tasks/fot-addpath.html).
        * You may need to reboot to reload the PATH before the command is found in the terminal.
        
Usage
-----

The following arguments are supported. Run `golivereload --help` for the most up-to-date help.

```
-cmd string
    Command to run after change is detected in files matching pattern. See include-pattern.
    (example: **/*.{scss} ./build.sh)
-debug
    Show debug output.
-delay int
    Delay this many milliseconds before before sending reload command.
-include string
    Only reload for files matching these patterns.
    Use ":" to separate patterns
    Use "**" (double star) to match multiple directories.
    Matches are relative to watched path.
    (default **/*.{html,shtml,tmpl,twig,xml,css,js,json}:**/*.{jpeg,jpg,gif,png,ico,cgi}:**/*.{php,py,pl,pm,rb})
-path string
    The directory to watch for changes.
    (default: current directory)
-port int
    Port to serve on.
    (default 35729)
-serve
    Start local webserver that serves files at -path.
-version
    Show golivereload version.
```

Attribution
-----------

Golivereload wouldn't be possible without:
* [Livereload.js](https://github.com/livereload/livereload-js)
* [notify](https://github.com/rjeczalik/notify)
