package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
)

import "html/template"

// para validación

// definimos nuestras estructuras de datos, una wiki tiene un titulo y un body.
// esta estructura define como será guardada nuestra página en memoria

// Page estructura para guardar las páginas en memoria
type Page struct {
	Title string
	Body  []byte // un slice de bytes , esto es mejor q un string, puesto q les lo q espera las librerias io
}

// variable GLOBAL templates, q inicializamos y cargamos las plantillas , Must es un contenedor q entra en pánico cuando hay un error, momento para salir del programa
// y devuelve un puntero *Templae inalterado.
var templates = template.Must(template.ParseFiles("edit.html", "view.html")) // si queremos añadir más plantillas lo unico q tenemos q hacer es añadir más argumnentos

// creamos otra variable global para manejar la validación.
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// esta función MustCompile parseará y compilará una expresión regular,  devuelve un regexp.Regexp.MustCompile si es distinto del compilador, y devuelve un erro como segundo parámetro.

//Para manejar la persistencia, nosotros creamos una función save()
// como argumento utiliza p, es decir un puntero a Page.
// devuelve un valor de tipo error , xq este es el tipo q devuelve el WriteFile
// si todo va bien, devuelve nil
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600) // este valor octal 0600, indica q el fichero debe ser creado con permisos de lectura-escritura
}

//además de grabar página, tb tenemos q cargarlas
//La función loadPage construye el nombre del fichero del título, y el contenido lo almacena en body.

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	// body, _ := ioutil.ReadFile(filename) // esta función devuelve un []byte y error. Nosotros no estamos manejando el error.
	// el identificador blanco _ es usado para lanzar el error
	// linea modificado para gesionar el error
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

// creamos el viewHandler para q los usuarios vean la vista de la wiki, y maneja prefijos con /view/
// el manejo de error en cada handler es un coñazo xq introduce código repetitvo, vamos centralizarlo, resecribimos la función y le ponemos otro parámetro, title string
// vamos a crear una función ANÓNIMA, go las llama funcion literals, makeHandler q se encargue de chequear los errores.
//una funcion anónima puede ser asingnada a una variable o invocada directamente.
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	// sustituimos la línea de abajo por la función para obtener el título
	//title := r.URL.Path[len("/view/"):]
	/* lo quitamos xq utilizamos el makeHandler
	  title, err := getTitle(w, r)
		if err != nil {
			return
		}
	*/
	// modificamos la línea de código siguiente, para introducir el tratamiento del error, y si la página no existe, hacer un redicet a editar.
	// p, _ := loadPage(title)
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound) // añade una constante, el código de http.StatusFound (302)
		return
	}
	// fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
	/* lo metemos en renderTemplate
	     t, _ := template.ParseFiles("view.html") // utilizamos la template view.html
	   	t.Execute(w, p)
	*/

	renderTemplate(w, "view", p)
}

//la función editHandler lo primero q hace es carga la página (si no existe, crea una structura vacía de Page), y visualiza un formulario HTML
func editHandler(w http.ResponseWriter, r *http.Request, title string) {

	//title := r.URL.Path[len("/edit/"):]
	/* lo quitamos xq utilizamos el makeHandler
	  title, err := getTitle(w, r)
		if err != nil {
			return
		}
	*/
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	/* utilizamos una template para evitar poner el código siguiente
	     fmt.Fprintf(w, "<h1>Editing %s</h1>"+
	   		"<form action=\"/save/%s\" method=\"POST\">"+
	   		"<textarea name=\"body\">%s</textarea><br>"+
	   		"<input type=\"submit\" value=\"Save\">"+
	   		"</form>",
	   		p.Title, p.Title, p.Body)
	*/
	// usando la template edit.HTML
	/* metemos lo siguiente en una función, renderTemplate
	  t, _ := template.ParseFiles("edit.html")
		t.Execute(w, p) // este método escribe el html generado en la variable w, es decir, en el ResponseWriter, los identificadores .Title y .Body, se refieren a p.Title y p.Body
	*/
	/* notas sobre templates:
	las directivas templates van entre doble llave. Y printf "%s" .Body , es una instrucción q llama a la salida .Body, pero como string en lugar de como stream de bytes, es lo mismo
	q fmt.Printf
	*/
	renderTemplate(w, "edit", p)
}

// función saveHandler
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	//title := r.URL.Path[len("/save/"):]
	/* lo quitamos xq utilizamos el makeHandler
	  title, err := getTitle(w, r)
		if err != nil {
			return
		}
	*/
	body := r.FormValue("body")                  // este valor es de tipo string
	p := &Page{Title: title, Body: []byte(body)} // hacemos un casting a []byte para transformar el string del body

	// err := p.save()                              // lo llama para escribir los datos al fichero title.txt
	err := p.save() // tengo q quitarle los :, xq no estoy creando una nueva variable error, ya la he creado en la primera línea; OJO SE LOS VUELVO A PONER XQ COMENTÉ LAS LÍNEAS DE ARRIBA
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

/*
renderTemplate llama a ParseFiles cada vez q una página es rendirizada. Podemos llamar a ParseFiles solamente una vez, en la inicializaicón del programa., y depues
pasear todas las plantillas en un simople puntero a template *Template , y usar un metodo ExecuteTemplate para renderizar una plantilla específica.
lo 1º q tenemos q hacer es crear una variable globar templates, e inicializarla con ParseFiles
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
*/
// como hemos utlizado el mismo código para plantillas en ambos handler, lo q hacemos es crear una función con este código común.
// modificamos renderTemplate para utilizar la variable global templates y llamar a templates.ExecuteTemplate con el valor de la plantilla apropiada.
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	//modificacion comentada arriba
	// t, err := template.ParseFiles(tmp + ".html")
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // esta función envía un código http específica del error, en este caso   "Internal Server Error", y el mensaje de error

	}
	/* lineas quitadas por la modificación de la variable global
	     err = t.Execute(w, p)
	   	if err != nil {
	   		http.Error(w, err.Error(), http.StatusInternalServerError)
	   	}
	*/
}

//Función q utilizar el validPath expresión para validar el path y extraer el título de la página
// si el título es válido , lo devuelve junto con un error nulo, y si no es válido, la funcion escribe un error 404,  devuelve
// el error para q sea manejado.
// parar crear un nuevo error, hemos de importar el paquete "errors"
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title") // lanza un nuevo error
	}
	return m[2], nil // The title is the second subexpression.
}

// funcion anónima, para centralizar el tratamiento de errores.
// es una función q coge un http.ResponseWriter y un http.Request , en otras palabras una http.HandlerFunc, extrae el título del path, lo valida con regexp,
// si el título no es válido, escribe un error en el ResponseWriter usndo la función http.NotFound
// si el título es válido, lo encapsula en la función fn, q tiene como argumentos, el ResponseWriter, el Request y el título.
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Here we will extract the page title from the Request,
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		// and call the provided handler 'fn'
		fn(w, r, m[2])
	}
}

func main() {
	/* p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page.")}
	   p1.save()
	   p2, _ := loadPage("TestPage")
	   fmt.Println(string(p2.Body))
	*/

	// Para usar nuestro handler, inicializamos http usando el viewHandler
	// http.HandleFunc("/view/", viewHandler) // esta función le dice al paquete http q maneja las peticiones de /view con el handler viewHandler

	// creamos un handler para editar
	// http.HandleFunc("/edit/", editHandler)
	// creamos un handler para guardar
	// http.HandleFunc("/save/", saveHandler)

	// usando el makeHandler
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	http.ListenAndServe(":8080", nil) // escucha por el puerto 8080 cualquier interface.

}
