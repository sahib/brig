Hugo Scroll
=
A live example of this theme is running at this site [hugoscroll.fredrikloch.me](http://hugoscroll.fredrikloch.me)
Using scroll
-
To get started with hugo you first need to download the [binaries](http://gohugo.io), whith these in place it is easy to get started.

    # Init site
    hugo new site My_New_Cool_Venture
    cd My_New_Cool_Venture

    # Get this cool theme
    git clone https://github.com/SenjinDarashiva/hugoscroll themes/hugoscroll

    # Start the watching your folder
    hugo --buildDrafts --theme="hugoscroll" --watch server

With this done you can start creating posts, for this theme theme there is some specific information that needs to be added to your
post header. For the first and the last post you need to add a class definition, the headers for this page look like this:

    +++
    title = "What is this"
    description = "First post"
    weight = 1
    type = "post"
    class="post first"
    +++

    +++
    title = "Finaly!"
    description = "Last Post"
    weight = 100
    type = "post"
    class="post last"
    +++

Every standard post must contain a weight between the weight of the first and the last to ensure correct ordering, in this case this
allows us to use any number between 2 -- 99

Site config
-
Apart from the regular config you can specify the following parameters to get extra features in the theme

    [Params]
      github = "Senjindarashiva"
      bitbucket = "floch"
      flickr = "senjin"
      twitter = "senjindarshiva"
      email = "fredrik.loch@outlook.com"
      description = ""
      linkedin = "fredrikloch"
      cover = "/images/background-cover.jpg"
      logo = "/img/logo-1.jpg"

Developing hugoscroll
=
In order to develop or make changes to the theme you will need to have the sass compiler and bourbon both installed.

To check installation run the following commands from a terminal and you should see the `> cli output` but your version numbers may vary.

#### SASS
```bash
sass -v
> Sass 3.3.4 (Maptastic Maple)
```
If for some reason SASS isn't installed follow the instructions from the [Sass install page](http://sass-lang.com/install)

#### Bourbon
```bash
bourbon help
> Bourbon 3.1.8
```
If Bourbon isn't installed follow the installation instructions on the [Bourbon website](http://bourbon.io)

Once installation is verified we will need to go mount the bourbon mixins into the `scss` folder.

From the project root run `bourbon install` with the correct path
```bash
bourbon install --path static/scss
> bourbon files installed to static/scss/bourbon/
```

Now that we have the bourbon mixins inside of the `scss` src folder we can now use the sass cli command to watch the scss files for changes and recompile them.

```bash
sass --watch static/scss:static/css
>>>> Sass is watching for changes. Press Ctrl-C to stop.
```

To minify the css files use the following command in the statics folder

```bash
curl -X POST -s --data-urlencode 'input@css/base.css' http://cssminifier.com/raw > css/base.min.css
```

Font-awesome icons
-
For more information on available icons: [font-awesome](http://fortawesome.github.io/Font-Awesome/)
The files supplied with this theme have a minor alteration to work around an issue with adblocks and social icons.
The changes means that the following classes is used:

* fa-tt -- twitter
* fa-fb -- facebook
* fa-ll -- linkedin
