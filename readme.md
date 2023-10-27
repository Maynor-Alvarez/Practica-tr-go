## Práctica guía de capacitación Back-End

Se creo una API que centraliza otros servicios de búsqueda canciones,
Como lo son:
- ITunes
  - https://itunes.apple.com/search?term=jack+johnson
  - http://api.chartlyrics.com/apiv1.asmx/SearchLyric?artist=string&song=string

el cual se le pueden mandar 3 parametros (name, artist, album) para realizar la busqueda en ambos EndPoints
Se consolidan las respuestas de estos con el mismo formato de salida.

### Instalación local

1. Tener Instalado Golang en su version 1.21
2. Instalar Docker en su version mas reciente
3. Clonar el repositorio
4. Tener instalado PostMan
5. Generar un contenedor mySql para la base de datos
   - docker run -d -p 33060:3306 --name mysql -e MYSQL_ROOT_PASSWORD=password mysql
6. Abrir una terminal y ubicarnos en la carpeta del proyecto y ejecutar los siguentes comandos:
   - go mod tidy
   - go mod download
   - go install
   - go build
   - ./Practica

Con estos comandos descargaremos las librerias y dependencias necesarias para el proyecto y generamos el ejecutable para el mismo

7. Ingresamos a postman, abrimos una nueva pestaña y colocamos la siguente URL:
   - Generamos el Token con: http://localhost:8080/token y lo copiamos
8. Abrimos una nueva pestaña y apuntamos a la siguente URL:
   - http://localhost:8080/songs?name=numb&artist=linkin park&album=meteora
Los parametros podremos cambiarlos según la busqueda de querramos realizar
   
y Asi podremos obtener un listado de las canciones de iTunes y ChartLyrics segun los filtros solicitados

### Base de datos

La base de datos es MySql y utiliza las siguentes credenciales
- Usuario: root
- Contraseña: password
- Puerto: 33060

Se pueden cambiar estas según sea necesario si se desea, con el comando mostrado en el apartado de "Instalación Local"

