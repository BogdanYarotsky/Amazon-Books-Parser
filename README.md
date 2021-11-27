# Amazon-Books-Parser
#### Video Demo:  https://www.youtube.com/watch?v=thR75861Btw
#### Description: Web application to find best books based on your interests

## Try it yourself

❗ [App hosted on Heroku](https://toppler.herokuapp.com/) ❗
Please, note that I'm not yet aware of how to use background threads for responsive web UIs. Plus, Ajax seems like an advanced topic for me. For this reason - load time at the moment is about 40-50 seconds. Please, be patient - it will load eventually.

## Application Features

- Find the best books from Amazon and Goodreads based on real people reviews!
- Do this in a single search query - no manual searching. Patented (actually not) sorting algorithm!
- Backend is written in Golang, because faster than Python and has easy-to-grasp concurrency support.
- Uses Headless Chrome to parse webpages and avoid captchas (no APIs from websites were available when I started)
- Aggregates results in the pretty CSS-styled table (Bootstrap FTW)

## Info about source code
- Static files folder contains CSS for the web pages styling
- Templates folder is used for HTML files which will be manipulated by Golang
- main.go contains boilerplate to handle web requests and call "GetBooks()" method when it is time
- book_parser.go is used to set up Headless Chrome instance via chromedp library. It is a beautiful wrapper over Chrome Dev Tools. Lib itself is alternative for Selenium, which is only available in Go
- This headless chrome is used as a crawler to get web pages from Amazon and Goodreads. Just for the HTML, we don't parse the pages yet, rather save it as a big string for further parsing.
- This HTML is then parsed with a standard library Go parser which is based on the Depth-First search algorithm.
- Every parsed page generates a fair amount of "Book" structs, which then will be sorted by specific website function (ex. SortAmazonBooks()) to find the best ones
- This list will be passed to results.html template for rendering


## Build the app without Heroku
There are two ways to use the app on the PC (Little programming experience is required)
### 1. With Go
1) Get the repo via ``` git clone https://github.com/BogdanYarotsky/Amazon-Books-Parser.git ```
2) Download Golang: https://golang.google.cn/doc/install
3) After setting up the Go - run command ```  go run .  ``` in the root folder 
4) If on Windows - Allow Access for Firewall in the pop-out window
5) Navigate to http://localhost:8080/ in your favorite browser
6) Look for books you are interested in (ex. queries: "Abraham Lincoln", "Python")

### 2. With Docker
1) Get the repo via ``` git clone https://github.com/BogdanYarotsky/Amazon-Books-Parser.git ```
2) Download Docker: https://www.docker.com/products/docker-desktop
3) After setting up Docker - run command ``` docker build -t parser .  ``` in the root folder 
4) After the image was built - run command ```docker run -d -p 8080:8080 --init parser``` 
5) Navigate to http://localhost:8080/ in your favorite browser
6) Look for books you are interested in (ex. queries: "Abraham Lincoln", "Python")