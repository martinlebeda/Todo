package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DIR_MAYBE = "Maybe"
const DIR_TEMPLATES = "Templates"

var todoPath = ""
var editCommand = ""
var taskSuffix = ""
var noteSuffix = ""
var notePath = ""
var contexts []string

const DIRSEP = "~"
const DONEPREFIX = "x "

func main() {
	numbPtr := flag.Int("port", 39095, "port number")

	// path to inbox with default value
	inboxDefault := filepath.Join(os.Getenv("HOME"), "Todo")
	flag.StringVar(&todoPath, "path", inboxDefault, "path for executable scripts")

	editorDefault := os.Getenv("EDITOR")
	flag.StringVar(&editCommand, "editor", editorDefault, "editor for underlaying files")

	flag.StringVar(&taskSuffix, "tasksuffix", ".txt", "suffix for task files (default '.txt')")
	flag.StringVar(&noteSuffix, "notesuffix", ".md", "suffix for note files (default '.md')")
	flag.StringVar(&notePath, "notepath", "", "path for note files")

	flag.Parse()

	// check existence path
	_ = os.MkdirAll(todoPath, os.ModePerm)

	// set router
	router := gin.Default()

	// client side store
	store := sessions.NewCookieStore([]byte("wcvwrqwvbdtyjerh"))
	router.Use(sessions.Sessions("TodoServe", store))

	// service
	router.GET("/quit", quitServer)
	//router.StaticFile("/favicon.ico", "resources/favicon.ico")
	router.GET("/favicon.ico", func(c *gin.Context) {
		data, err := Asset("resources/favicon.ico")
		if err != nil {
			log.Print("resources/favicon.ico not found")
		}

		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write(data)
	})

	// API for external app
	router.POST("/insert", insertTodo) // for call from external (ie. browser extension)

	// special
	router.GET("/index", listTodo)
	router.POST("/search", search)
	router.POST("/clear", clear)
	router.POST("/move/:task/:dir", moveTask)

	// manage lists
	router.GET("/list/:path", listRender)
	router.POST("/list/:path/add", addTask)

	// manage tasks
	router.GET("/task/:task", taskSimpleRender)
	router.GET("/task/:task/full", taskFullRender)
	router.POST("/task/:task/done", taskDone)
	router.POST("/task/:task/prio/:prio", taskPrio)
	router.POST("/task/:task/context/:context", taskContext)
	router.GET("/task/:task/edit", taskEdit)
	router.GET("/task/:task/note", taskNote)
	router.DELETE("/task/:task/delete", taskDelete)

	// set listen port
	portNumberStr := strconv.Itoa(*numbPtr)
	router.Run(":" + portNumberStr)
}

func quitServer(c *gin.Context) {
	c.String(http.StatusOK, "quiting...")
	go func() {
		fmt.Println("quiting launcher...")
		time.Sleep(time.Millisecond * 50)
		os.Exit(0)
	}()
}

func insertTodo(c *gin.Context) {
	title := c.PostForm("title")
	todoUrl := c.PostForm("url")

	fmt.Println("data: " + title + " -> " + todoUrl)

	title = normalizeFileName(title)
	err := ioutil.WriteFile(filepath.Join(todoPath, "Inbox", title+".txt"), []byte(todoUrl), 0644)
	if err != nil {
		log.Println("Unable write file '" + title + ".txt'")
	}

	c.JSON(200, gin.H{
		"status": "posted",
	})
}

// normalize unvanted character
func normalizeFileName(title string) string {
	r := strings.NewReplacer("<", " ",
		">", " ",
		"|", " ",
		"/", " ",
		".", " ",
		":", " ",
		"?", " ",
		"!", " ",
		"\\", " ")
	nrmt := r.Replace(title)
	nrmt = strings.TrimSpace(nrmt)
	return nrmt
}

func listTodo(c *gin.Context) {
	body := getIndexList(c)

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(body))
}

func search(c *gin.Context) {
	search := c.PostForm("search")

	// store param to session
	session := sessions.Default(c)
	session.Set("search", search)
	session.Save()

	// send reply
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(listTodoDir(c, todoPath, search, false, "")))
}

func clear(c *gin.Context) {
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(listTodoDir(c, todoPath, "", true, "")))
}

func loadContexts() []string {
	f, err := os.Open(filepath.Join(todoPath, "contexts.txt"))
	if err != nil {
		log.Println("Unable to load contexts.txt")
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return lines
}

func getIndexList(c *gin.Context) string {
	contexts = loadContexts()

	// base of app
	body := `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1">
            <meta name="intercoolerjs:use-actual-http-method" content="true"/>
			<link rel="shortcut icon" href="/favicon.ico" type="image/x-icon">
			<link rel="icon" href="/favicon.ico" type="image/x-icon">
            <title>Todo</title>

            <script src="https://code.jquery.com/jquery-3.1.1.min.js"></script>
            <script src="https://cdn.jsdelivr.net/npm/jquery.hotkeys@0.1.0/jquery.hotkeys.min.js"></script>
			<!-- script src="http://nekman.github.io/keynavigator/keynavigator-min.js"></script -->
            <script src="https://intercoolerreleases-leaddynocom.netdna-ssl.com/intercooler-1.2.1.min.js"></script>
            <script defer src="https://use.fontawesome.com/releases/v5.0.3/js/all.js"></script>

            <!-- link rel="stylesheet" href="https://gitcdn.link/repo/Chalarangelo/mini.css/master/dist/mini-default.min.css" -->
            <link rel="stylesheet" href="https://gitcdn.link/repo/Chalarangelo/mini.css/master/dist/mini-dark.min.css">
            <style>
                .hiddencontrols {
                    visibility: hidden;
                }

                .dropdown {
                    position: relative;
                    display: inline-block;
                }

                .dropdown:hover .dropdown-content {
                    display: block;
                }

                /* Dropdown Content (Hidden by Default) */
                .dropdown-content {
                    display: none;
                    position: absolute;
                    background-color: #f9f9f9;
                    min-width: 160px;
                    box-shadow: 0px 8px 16px 0px rgba(0,0,0,0.2);
                    z-index: 1;
                }

				.nolink {
					text-decoration: none;
					color: #d0d0d0;f
				}

                /* Links inside the dropdown */
                .dropdown-content a {
                    color: black;
                    padding: 2px 6px;
                    text-decoration: none;
                    display: block;
                }

                /* Change color of dropdown links on hover */
                .dropdown-content a:hover {background-color: #a1a1a1}
            </style>
        </head>
        <body style="height: 100vh; display: flex; flex-direction: column;">
        <header style="flex-grow: 1; min-height: 44px; display: flex;">
            <input id="searchinput" type="text" style=" padding: 5px; flex-grow: 999;" name="search" ic-post-to="/search" ic-trigger-on="keyup changed"
                           ic-trigger-delay="500ms" ic-target="main" placeholder="search (~regex) (+maybe)" value="` + getSessionSearch("", c) + `">
            <button type="button" id="searchclear" onclick='$("#searchinput").val(""); $("#searchinput").keyup();'><i class="fas fa-arrow-left"></i></button>
            <button type="button" id="search_a" onclick='$("#searchinput").val("(A) "); $("#searchinput").keyup();'>(A)</button>
            <button type="button" id="search_a" onclick='$("#searchinput").val("~^\([AB]\) "); $("#searchinput").keyup();'>(B)</button>
            <button type="button" id="search_a" onclick='$("#searchinput").val("~^\([ABC]\) "); $("#searchinput").keyup();'>(C)</button>
            <button type="button" id="clearrepo" ic-post-to="/clear" ic-target="main"><i class="fas fa-trash-alt"></i></button>
        </header>
        <main style="flex-grow: 999; overflow-y: scroll;">
`
	body += listTodoDir(c, todoPath, "", false, "")
	body += `</main>
        </body>
		<script>
			// shortcuts
			$(document).bind('keydown', 'Alt+g', function(){ $("#searchinput").focus() });
			$(document).bind('keydown', 'Alt+Shift+g', function(){ $("#searchclear").click() });
			$(document).bind('keydown', 'Alt+Shift+a', function(){ $("#search_a").click() });

			// navigate
			$(document).ready(function() {
			  $('li').keynavigator();
			  // $('li').keynavigator(/* optional settings */);
			});
			$('li').keynavigator();
		</script>
        </html>`
	return body
}

/*
"esc","tab","space","return","backspace","scroll","capslock","numlock","insert","home","del","end","pageup","pagedown",
                    "left","up","right","down",
                    "f1","f2","f3","f4","f5","f6","f7","f8","f9","f10","f11","f12",
                    "1","2","3","4","5","6","7","8","9","0",
                    "a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z",
                    "Ctrl+a","Ctrl+b","Ctrl+c","Ctrl+d","Ctrl+e","Ctrl+f","Ctrl+g","Ctrl+h","Ctrl+i","Ctrl+j","Ctrl+k","Ctrl+l","Ctrl+m",
                    "Ctrl+n","Ctrl+o","Ctrl+p","Ctrl+q","Ctrl+r","Ctrl+s","Ctrl+t","Ctrl+u","Ctrl+v","Ctrl+w","Ctrl+x","Ctrl+y","Ctrl+z",
                    "Shift+a","Shift+b","Shift+c","Shift+d","Shift+e","Shift+f","Shift+g","Shift+h","Shift+i","Shift+j","Shift+k","Shift+l",
                    "Shift+m","Shift+n","Shift+o","Shift+p","Shift+q","Shift+r","Shift+s","Shift+t","Shift+u","Shift+v","Shift+w","Shift+x",
                    "Shift+y","Shift+z",
                    "Alt+a","Alt+b","Alt+c","Alt+d","Alt+e","Alt+f","Alt+g","Alt+h","Alt+i","Alt+j","Alt+k","Alt+l",
                    "Alt+m","Alt+n","Alt+o","Alt+p","Alt+q","Alt+r","Alt+s","Alt+t","Alt+u","Alt+v","Alt+w","Alt+x","Alt+y","Alt+z",
                    "Ctrl+esc","Ctrl+tab","Ctrl+space","Ctrl+return","Ctrl+backspace","Ctrl+scroll","Ctrl+capslock","Ctrl+numlock",
                    "Ctrl+insert","Ctrl+home","Ctrl+del","Ctrl+end","Ctrl+pageup","Ctrl+pagedown","Ctrl+left","Ctrl+up","Ctrl+right",
                    "Ctrl+down",
                    "Ctrl+f1","Ctrl+f2","Ctrl+f3","Ctrl+f4","Ctrl+f5","Ctrl+f6","Ctrl+f7","Ctrl+f8","Ctrl+f9","Ctrl+f10","Ctrl+f11","Ctrl+f12",
                    "Shift+esc","Shift+tab","Shift+space","Shift+return","Shift+backspace","Shift+scroll","Shift+capslock","Shift+numlock",
                    "Shift+insert","Shift+home","Shift+del","Shift+end","Shift+pageup","Shift+pagedown","Shift+left","Shift+up",
                    "Shift+right","Shift+down",
                    "Shift+f1","Shift+f2","Shift+f3","Shift+f4","Shift+f5","Shift+f6","Shift+f7","Shift+f8","Shift+f9","Shift+f10","Shift+f11","Shift+f12",
                    "Alt+esc","Alt+tab","Alt+space","Alt+return","Alt+backspace","Alt+scroll","Alt+capslock","Alt+numlock",
                    "Alt+insert","Alt+home","Alt+del","Alt+end","Alt+pageup","Alt+pagedown","Alt+left","Alt+up","Alt+right","Alt+down",
                    "Alt+f1","Alt+f2","Alt+f3","Alt+f4","Alt+f5","Alt+f6","Alt+f7","Alt+f8","Alt+f9","Alt+f10","Alt+f11","Alt+f12"
*/

func listTodoDir(c *gin.Context, dirName string, search string, clear bool, clientListId string) string {
	body := `<ul id="` + clientListId + `">`

	search = getSessionSearch(search, c)

	// walk directories
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		// remove invisible files and dirs
		if strings.HasPrefix(file.Name(), ".") ||
			(file.Name() == "contexts.txt") ||
			strings.HasSuffix(file.Name(), "~") ||
			IsIgnoreDir(file, search) {
			continue
		}

		dirList := filepath.Join(dirName, file.Name())
		if file.IsDir() {
			// if directory is ok for search -> all item is ok too
			subsearch := search
			if CheckFilterItem(search, file.Name()) {
				subsearch = ""
			}
			body += renderList(c, dirList, subsearch, clear)
		} else {
			// clear
			if clear && strings.HasPrefix(file.Name(), DONEPREFIX) {
				os.Remove(dirList)
				continue
			}

			// remove by filter
			itemOk := CheckFilterItem(search, file.Name())
			if !itemOk {
				continue
			}

			// render item
			body += renderSimpleItem(file.Name(), dirName)
		}
	}

	body += "</ul>\n"

	return body
}

// load search from session
func getSessionSearch(search string, c *gin.Context) string {
	if search == "" && c != nil {
		session := sessions.Default(c)
		s := session.Get("search")
		if s != nil {
			search = s.(string)
		}
	}
	return search
}

func IsIgnoreDir(file os.FileInfo, search string) bool {
	return (file.IsDir() && (file.Name() == DIR_TEMPLATES)) ||
		(!strings.Contains(search, "+maybe") && (file.IsDir() && (file.Name() == DIR_MAYBE)))
}

func CheckFilterItem(search string, fileName string) bool {
	itemOk := true

	re := regexp.MustCompile(" *\\+maybe *")
	search = re.ReplaceAllString(search, "")

	trimSpace := strings.TrimSpace(search)
	if len(trimSpace) > 0 {
		if strings.HasPrefix(search, "~") {
			// regex
			search = strings.TrimPrefix(search, "~")
			pattern := strings.Replace(search, "(", "\\(", -1)
			pattern = strings.Replace(pattern, ")", "\\)", -1)
			pattern = "(?i)" + pattern
			matched, _ := regexp.MatchString(pattern, fileName)
			if !matched {
				itemOk = false
			}
		} else {
			// simple search
			normSearch := NormalizeString(search)
			normName := NormalizeString(fileName)
			if normSearch != "" && !strings.Contains(normName, normSearch) {
				itemOk = false
			}
		}
	}
	return itemOk
}

func renderList(c *gin.Context, dirList string, search string, clear bool) string {
	base, style, listId, clientId := prepareList(dirList)
	body := `<li ic-src="/list/` + listId + `" ic-replace-target="true"><a class="nolink"><span style="` + style + `" 
						                    ic-post-to="/list/` + listId + `/add"        
						                    ic-prompt="New task in ` + base + `:" ic-include='{"cltgt": "` + clientId + `"}'
						                    >` + base + "</span></a>\n"
	body += listTodoDir(c, dirList, search, clear, clientId)
	body += "</li>\n"
	return body
}

func prepareList(dirList string) (string, string, string, string) {
	base := filepath.Base(dirList)
	style := "color: darkseagreen;"
	// set style for item
	if strings.HasPrefix(base, "(A) ") {
		style += "font-weight: bold;"
	} else {
		style += "font-size: small;"
		style += "font-style: italic;"
	}
	listId := encodeListId(dirList)
	clientId := uuid.Must(uuid.NewV4()).String()
	return base, style, listId, clientId
}

func encodeListId(dirList string) string {
	return url.PathEscape(strings.Replace(dirList, "/", DIRSEP, -1))
}

func renderContexts(fileName string) string {
	itemName := fileName
	for _, element := range contexts {
		itemName = strings.Replace(itemName, element, `<span style="color: darkcyan;" onclick='$("#searchinput").val("` + element + `"); $("#searchinput").keyup(); event.stopPropagation();'>`+
			element+ `</span>`, -1)
	}
    return itemName
}

func renderSimpleItem(fileName string, dirName string) string {
	style, taskId, clientId := prepareItem(fileName, dirName)

	renderedItem := `<li id="` + clientId + `"><a class="nolink" ic-get-from="/task/` + taskId + `/full" ic-target="#` + clientId + `" ic-replace-target="true">` +
		`<span style="` + style + `" >` + renderContexts(fileName) + `</span></a> ` +
		"</li>\n"
	return renderedItem
}

func renderFullItem(fileName string, dirName string) string {
	if len(contexts) == 0 { // reboot server?
		contexts = loadContexts()
	}

	style, taskId, clientId := prepareItem(fileName, dirName)

	content := ""
	fullFileName := filepath.Join(dirName, fileName)
	fileInfo, e := os.Stat(fullFileName)
	fmt.Println(e)
	if (e == nil) && (fileInfo.Size() > 0) && strings.HasSuffix(fileName, ".txt") {
		bytes, e := ioutil.ReadFile(fullFileName)
		if e == nil {
			output := blackfriday.Run(bytes)
			content = string(output[:])
		}
	}
	content = strings.Replace(content, "<a ", `<a target="_blank"`, -1)

	maybeDirname := strings.Replace(dirName, todoPath, DIR_MAYBE, -1)
	maybeIdPath := url.PathEscape(strings.Replace(maybeDirname, "/", DIRSEP, -1))

	renderedItem := `<li id="` + clientId + `"><a class="nolink" ic-get-from="/task/` + taskId + `" ic-target="#` + clientId + `" ic-replace-target="true">
                        <span style="` + style + `">` + renderContexts(fileName) + `</span> ` +
		`<div>
            <button type="button" ic-post-to="/task/` + taskId + `/done" ic-target="#` + clientId + `" ic-replace-target="true"><i class="fas fa-check"></i></button>
            <button type="button" ic-get-from="/task/` + taskId + `/edit"><i class="fas fa-pencil-alt"></i></button>
            <button type="button" ic-get-from="/task/` + taskId + `/note"><i class="far fa-edit"></i></button>
            <button type="button" ic-post-to="/task/` + taskId + `/prio/A" ic-target="#` + clientId + `" ic-replace-target="true">(A)</button>
            <button type="button" ic-post-to="/task/` + taskId + `/prio/B" ic-target="#` + clientId + `" ic-replace-target="true">(B)</button>
            <button type="button" ic-post-to="/task/` + taskId + `/prio/C" ic-target="#` + clientId + `" ic-replace-target="true">(C)</button>
            <button type="button" ic-post-to="/move/` + taskId + `/` + maybeIdPath + `" ic-target="main"><i class="fas fa-archive"></i></button>
            <button type="button" ic-delete-from="/task/` + taskId + `/delete" ic-target="#` + clientId + `"><i class="fas fa-trash-alt"></i></a>
           </div>` + content +
		`<div>` + getMoveDirs(taskId) + `</div>` +
		`<div>` + getContexts(taskId, clientId) + `</div>` +
		"</li>\n"
	return renderedItem
}

func prepareItem(fileName string, dirName string) (string, string, string) {
	style := ""
	// set style for item
	if strings.HasPrefix(fileName, DONEPREFIX) {
		style += "font-weight: normal; color: #36454c; text-decoration: line-through;" // silver
	} else if strings.HasPrefix(fileName, "(A) ") {
		style += "font-weight: bold;"
	}
	// add item to result
	taskId := url.PathEscape(strings.Replace(filepath.Join(dirName, fileName), "/", DIRSEP, -1))
	clientId := uuid.Must(uuid.NewV4()).String()
	return style, taskId, clientId
}

func getMoveDirs(taskId string) string {
	result := ""
	for _, element := range GetDirectories() {
		element = strings.TrimPrefix(element, todoPath)
		if (strings.TrimSpace(element) == "") || strings.HasPrefix(element, "/"+DIR_MAYBE) || strings.HasPrefix(element, "/"+DIR_TEMPLATES) {
			continue
		}
		pathEscape := encodeListId(element)
		result += `<button type="button" class="small" ic-post-to="/move/` + taskId + `/` + pathEscape + `" ic-target="main">` + element[1:] + `/</button>`
		// element is the element from someSlice for where we are
	}
	return result
}

func getContexts(taskId string, clientId string) string {
	result := ""
	for _, element := range contexts {
		//element = strings.TrimPrefix(element, todoPath)
		//if (strings.TrimSpace(element) == "") || strings.HasPrefix(element, "/"+DIR_MAYBE) || strings.HasPrefix(element, "/"+DIR_TEMPLATES) {
		//continue
		//}context
		//pathEscape :=  encodeListId(element)
		///task/:task/context/:context
		result += `<button type="button" class="small" ic-post-to="/task/` + taskId + `/context/` + element +
			`" ic-target="#` + clientId + `" ic-replace-target="true">` + element + `</button>`
		// element is the element from someSlice for where we are
	}
	return result
}

func getTaskFromUrlPath(c *gin.Context) (string, string) {
	task, err := url.PathUnescape(c.Param("task"))
	if err != nil {
		log.Println("Unable render task '" + task)
	}
	task = strings.Replace(task, DIRSEP, "/", -1)
	dir := filepath.Dir(task)
	name := filepath.Base(task)
	return dir, name
}

func taskSimpleRender(c *gin.Context) {
	dir, name := getTaskFromUrlPath(c)

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(renderSimpleItem(name, dir)))
}

func taskFullRender(c *gin.Context) {
	dir, name := getTaskFromUrlPath(c)

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(renderFullItem(name, dir)))
}

func taskDone(c *gin.Context) {
	dir, oldName := getTaskFromUrlPath(c)

	// make new task name
	var newName string
	if strings.HasPrefix(oldName, DONEPREFIX) {
		newName = strings.TrimPrefix(oldName, DONEPREFIX)
	} else {
		newName = DONEPREFIX + RemoveAllTags(oldName)
	}

	// rename file
	err := os.Rename(filepath.Join(dir, oldName), filepath.Join(dir, newName))
	if err != nil {
		fmt.Println(err)
	}

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(renderSimpleItem(newName, dir)))
}

func taskPrio(c *gin.Context) {
	dir, oldName := getTaskFromUrlPath(c)

	prio := c.Param("prio")

	// make new task name
	var newName string
	donePrefix := "(" + prio + ") "
	if strings.HasPrefix(oldName, donePrefix) {
		newName = strings.TrimPrefix(oldName, donePrefix)
	} else {
		newName = donePrefix + RemoveAllPrio(oldName)
	}

	// rename file
	err := os.Rename(filepath.Join(dir, oldName), filepath.Join(dir, newName))
	if err != nil {
		fmt.Println(err)
	}

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(renderSimpleItem(newName, dir)))
}

func taskContext(c *gin.Context) {
	dir, oldName := getTaskFromUrlPath(c)

	context := c.Param("context")

	// make new task name
	var newName string
	if strings.Contains(oldName, context) {
		newName = strings.Replace(oldName, " "+context, "", -1)
	} else {
		ext := filepath.Ext(oldName)
		basename := strings.TrimSuffix(oldName, ext)
		newName = basename + " " + context + ext
	}

	// rename file
	err := os.Rename(filepath.Join(dir, oldName), filepath.Join(dir, newName))
	if err != nil {
		fmt.Println(err)
	}

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(renderSimpleItem(newName, dir)))
}

func RemoveAllPrio(name string) string {
	r, _ := regexp.Compile("^\\(.\\) ")
	return r.ReplaceAllString(name, "")
}

func RemoveAllTags(name string) string {
	wkName := RemoveAllPrio(name)

	for _, element := range contexts {
        wkName = strings.Replace(wkName, element, "", -1)
	}

	r, _ := regexp.Compile(" +")
	wkName = r.ReplaceAllString(wkName, " ")

	wkName = strings.Replace(wkName, " .", ".", -1)

	return wkName
}

func taskEdit(c *gin.Context) {
	dir, name := getTaskFromUrlPath(c)

	cmd := exec.Command(editCommand, filepath.Join(dir, name))
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	c.Writer.WriteHeader(http.StatusOK)
}

func taskNote(c *gin.Context) {
	_, name := getTaskFromUrlPath(c)

	name = RemoveAllTags(name)
	name = strings.TrimSuffix(name, taskSuffix)
	name = name + noteSuffix

    path := filepath.Join(notePath, name)
    cmd := exec.Command(editCommand, path)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	c.Writer.WriteHeader(http.StatusOK)
}

func moveTask(c *gin.Context) {
	dir, name := getTaskFromUrlPath(c)

	// decode target directory
	tgtDir, err := url.PathUnescape(c.Param("dir"))
	if err != nil {
		log.Println("Unable render task '" + tgtDir)
		return
	}
	tgtDir = strings.Replace(tgtDir, DIRSEP, "/", -1)
	tgtDir = filepath.Join(todoPath, tgtDir)

	// create dir if not exists
	err = os.MkdirAll(tgtDir, os.ModePerm)
	if err != nil {
		log.Println("Unable create target directory: " + tgtDir)
		return
	}

	// do rename
	errRename := os.Rename(filepath.Join(dir, name), filepath.Join(tgtDir, name))
	c.Writer.Header().Add("X-IC-Refresh", "/list/"+encodeListId(tgtDir)+","+"/list/"+encodeListId(dir))
	if errRename != nil {
		log.Fatal(errRename)
	}

	// make refresh index
	c.Writer.WriteHeader(http.StatusOK)
	//c.Writer.Header().Add("X-IC-Remove", "true")
	//c.Writer.Write([]byte(listTodoDir(todoPath, "", false, "")))
}

func listRender(c *gin.Context) {
	path, err := decodePath(c)

	if err != nil {
		log.Println("Unable render list")
		return
	}

	//cltgt := c.PostForm("cltgt")

	// make refresh index
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte(renderList(c, path, "", false)))
}

func decodePath(c *gin.Context) (string, error) {
	path, err := url.PathUnescape(c.Param("path"))
	if err != nil {
		log.Println("Unable add task")
	}
	path = strings.Replace(path, DIRSEP, "/", -1)
	return path, err
}

func addTask(c *gin.Context) {
	path, err := decodePath(c)

	//cltgt := c.PostForm("cltgt")
	name := c.PostForm("ic-prompt-value")

	if name == "-" {
		listDeleteIfEmpty(filepath.Join(path))
		parentDir := filepath.Dir(path)
		c.Writer.Header().Add("X-IC-Refresh", "/list/"+encodeListId(parentDir))
	} else if strings.HasPrefix(name, "/") {
		// create directory
		name = strings.TrimSpace(name)
		err = os.Mkdir(filepath.Join(path, normalizeFileName(name)), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// create file
		_, err = os.Create(filepath.Join(path, normalizeFileName(name)+".txt"))
		if err != nil {
			log.Fatal(err)
		}
	}

	// make refresh index

	c.Writer.WriteHeader(http.StatusOK)
	//c.Writer.Write([]byte(listTodoDir(path, "", false, cltgt)))
}

func taskDelete(c *gin.Context) {
	dir, name := getTaskFromUrlPath(c)

	err := os.Remove(filepath.Join(dir, name))
	if err != nil {
		log.Fatal(err)
	}

	listDeleteIfEmpty(filepath.Join(dir))

	parentDir := filepath.Dir(dir)
	c.Writer.Header().Add("X-IC-Refresh", "/list/"+encodeListId(parentDir))
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Header().Add("X-IC-Remove", "true")
}

func listDeleteIfEmpty(dir string) {
	isEmpty, err := IsEmpty(dir)
	if err != nil {
		log.Fatal(err)
	}
	if isEmpty && !isRoot(dir, todoPath) {
		err = os.Remove(dir)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func isRoot(dir string, root string) bool {
    return filepath.Dir(dir) == root
}

// Returns the names of the subdirectories (including their paths)
// that match the specified search pattern in the specified directory.
func GetDirectories() []string {
	dirs := make([]string, 0, 144)
	filepath.Walk(todoPath, func(path string, fi os.FileInfo, err error) error {
		if !fi.IsDir() || strings.HasPrefix(fi.Name(), ".") || strings.Contains(path, "/.") {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	return dirs
}
