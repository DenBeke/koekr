Koekr
=====

Static site generator for HTLM templates.

Why?
----

I needed a tool to quickly generate multiple static HTML pages with the same header/footer/navigation.  
I didn't need a complete markdown-based static site generator like Hugo or Jykell, but just some simple HTML themes.

So if you're expecting a full-grown magic framework to write your blog, take a look at 'real things'.

I only created Koekr to help me create static websites for clients.


Usage
-----

1. Create an `config.toml` file and add some variables to it.  
Example of config:

        site_title = "Koekr"
        
        [copyright]
        
        name = "DenBeke"
        year = 2017
    

1. Create HTML files in the `pages` directory.
The complete content of those files can be accessed in `index.html` like this: `{{ .page.content }}`
If you want to add variables to your pages, you can do so by writing TOML, enclosed by `---`.  
Example of a page:

        ---
        title = "Hello world!"
        ---
        <p>Koekestad!</p>

1. Create an `index.html` page and use the variables you defined.  
Example of the index template:

        <!DOCTYPE html>
        <html>
        <head>
            <title>{{ .site_title }}</title>
        </head>
        <body>
        
            <h1>{{ .page.title }}</h1>
        
            {{.page.content}}
        
            <footer>
                &copy; {{ .copyright.year }} - {{ .copyright.name }}
            </footer>
        
        </body>
        </html>

1. Put your assets in the `assets` directory. That directory gets copied to `generated/assets` each time you run Koekr.

1. Run Koekr from the command line:
        
        $ koekr

1. Profit!


Author
------

[Mathias Beke](https://denbeke.be)