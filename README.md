# TP0: Docker + Comunicaciones + Concurrencia

## Índice

- [Instrucciones de uso](#instrucciones-de-uso)
  - [Servidor](#servidor)
  - [Cliente](#cliente)
  - [Ejemplo](#ejemplo)
- [Parte 1: Introducción a Docker](#parte-1-introduccion-a-docker)
  - [Ejercicio N°1](#ejercicio-n1)
  - [Ejercicio N°2](#ejercicio-n2)
  - [Ejercicio N°3](#ejercicio-n3)
  - [Ejercicio N°4](#ejercicio-n4)
- [Parte 2: Repaso de Comunicaciones](#parte-2-repaso-de-comunicaciones)
  - [Ejercicio N°5](#ejercicio-n5)
  - [Ejercicio N°6](#ejercicio-n6)
  - [Ejercicio N°7](#ejercicio-n7)
- [Parte 3: Repaso de Concurrencia](#parte-3-repaso-de-concurrencia)
  - [Ejercicio N°8](#ejercicio-n8)
- [Condiciones de Entrega](#condiciones-de-entrega)
- [Resolución de Ejercicios](#resolucion-de-ejercicios)
  - [Ejercicio 1](#ejercicio-1)
  - [Ejercicio 2](#ejercicio-2)
  - [Ejercicio 3](#ejercicio-3)
  - [Ejercicio 4](#ejercicio-4)
  - [Ejercicio 5](#ejercicio-5)
  - [Ejercicio 6](#ejercicio-6)
  - [Ejercicio 7](#ejercicio-7)
  - [Ejercicio 8](#ejercicio-8)

En el presente repositorio se provee un esqueleto básico de cliente/servidor, en donde todas las dependencias del mismo se encuentran encapsuladas en containers. Los alumnos deberán resolver una guía de ejercicios incrementales, teniendo en cuenta las condiciones de entrega descritas al final de este enunciado.

 El cliente (Golang) y el servidor (Python) fueron desarrollados en diferentes lenguajes simplemente para mostrar cómo dos lenguajes de programación pueden convivir en el mismo proyecto con la ayuda de containers, en este caso utilizando [Docker Compose](https://docs.docker.com/compose/).

## Instrucciones de uso
El repositorio cuenta con un **Makefile** que incluye distintos comandos en forma de targets. Los targets se ejecutan mediante la invocación de:  **make \<target\>**. Los target imprescindibles para iniciar y detener el sistema son **docker-compose-up** y **docker-compose-down**, siendo los restantes targets de utilidad para el proceso de depuración.

Los targets disponibles son:

| target  | accion  |
|---|---|
|  `docker-compose-up`  | Inicializa el ambiente de desarrollo. Construye las imágenes del cliente y el servidor, inicializa los recursos a utilizar (volúmenes, redes, etc) e inicia los propios containers. |
| `docker-compose-down`  | Ejecuta `docker-compose stop` para detener los containers asociados al compose y luego  `docker-compose down` para destruir todos los recursos asociados al proyecto que fueron inicializados. Se recomienda ejecutar este comando al finalizar cada ejecución para evitar que el disco de la máquina host se llene de versiones de desarrollo y recursos sin liberar. |
|  `docker-compose-logs` | Permite ver los logs actuales del proyecto. Acompañar con `grep` para lograr ver mensajes de una aplicación específica dentro del compose. |
| `docker-image`  | Construye las imágenes a ser utilizadas tanto en el servidor como en el cliente. Este target es utilizado por **docker-compose-up**, por lo cual se lo puede utilizar para probar nuevos cambios en las imágenes antes de arrancar el proyecto. |
| `build` | Compila la aplicación cliente para ejecución en el _host_ en lugar de en Docker. De este modo la compilación es mucho más veloz, pero requiere contar con todo el entorno de Golang y Python instalados en la máquina _host_. |

### Servidor

Se trata de un "echo server", en donde los mensajes recibidos por el cliente se responden inmediatamente y sin alterar. 

Se ejecutan en bucle las siguientes etapas:

1. Servidor acepta una nueva conexión.
2. Servidor recibe mensaje del cliente y procede a responder el mismo.
3. Servidor desconecta al cliente.
4. Servidor retorna al paso 1.


### Cliente
 se conecta reiteradas veces al servidor y envía mensajes de la siguiente forma:
 
1. Cliente se conecta al servidor.
2. Cliente genera mensaje incremental.
3. Cliente envía mensaje al servidor y espera mensaje de respuesta.
4. Servidor responde al mensaje.
5. Servidor desconecta al cliente.
6. Cliente verifica si aún debe enviar un mensaje y si es así, vuelve al paso 2.

### Ejemplo

Al ejecutar el comando `make docker-compose-up`  y luego  `make docker-compose-logs`, se observan los siguientes logs:

```
client1  | 2024-08-21 22:11:15 INFO     action: config | result: success | client_id: 1 | server_address: server:12345 | loop_amount: 5 | loop_period: 5s | log_level: DEBUG
client1  | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:14 DEBUG    action: config | result: success | port: 12345 | listen_backlog: 5 | logging_level: DEBUG
server   | 2024-08-21 22:11:14 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°3
client1  | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°3
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°5
client1  | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°5
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:40 INFO     action: loop_finished | result: success | client_id: 1
client1 exited with code 0
```


## Parte 1: Introducción a Docker
En esta primera parte del trabajo práctico se plantean una serie de ejercicios que sirven para introducir las herramientas básicas de Docker que se utilizarán a lo largo de la materia. El entendimiento de las mismas será crucial para el desarrollo de los próximos TPs.

### Ejercicio N°1:
Definir un script de bash `generar-compose.sh` que permita crear una definición de Docker Compose con una cantidad configurable de clientes.  El nombre de los containers deberá seguir el formato propuesto: client1, client2, client3, etc. 

El script deberá ubicarse en la raíz del proyecto y recibirá por parámetro el nombre del archivo de salida y la cantidad de clientes esperados:

`./generar-compose.sh docker-compose-dev.yaml 5`

Considerar que en el contenido del script pueden invocar un subscript de Go o Python:

```
#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"
python3 mi-generador.py $1 $2
```

En el archivo de Docker Compose de salida se pueden definir volúmenes, variables de entorno y redes con libertad, pero recordar actualizar este script cuando se modifiquen tales definiciones en los sucesivos ejercicios.

### Ejercicio N°2:
Modificar el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) debe ser inyectada en el container y persistida por fuera de la imagen (hint: `docker volumes`).


### Ejercicio N°3:
Crear un script de bash `validar-echo-server.sh` que permita verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo. Dado que el servidor es un echo server, se debe enviar un mensaje al servidor y esperar recibir el mismo mensaje enviado.

En caso de que la validación sea exitosa imprimir: `action: test_echo_server | result: success`, de lo contrario imprimir:`action: test_echo_server | result: fail`.

El script deberá ubicarse en la raíz del proyecto. Netcat no debe ser instalado en la máquina _host_ y no se pueden exponer puertos del servidor para realizar la comunicación (hint: `docker network`). `


### Ejercicio N°4:
Modificar servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM. Terminar la aplicación de forma _graceful_ implica que todos los _file descriptors_ (entre los que se encuentran archivos, sockets, threads y procesos) deben cerrarse correctamente antes que el thread de la aplicación principal muera. Loguear mensajes en el cierre de cada recurso (hint: Verificar que hace el flag `-t` utilizado en el comando `docker compose down`).

## Parte 2: Repaso de Comunicaciones

Las secciones de repaso del trabajo práctico plantean un caso de uso denominado **Lotería Nacional**. Para la resolución de las mismas deberá utilizarse como base el código fuente provisto en la primera parte, con las modificaciones agregadas en el ejercicio 4.

### Ejercicio N°5:
Modificar la lógica de negocio tanto de los clientes como del servidor para nuestro nuevo caso de uso.

#### Cliente
Emulará a una _agencia de quiniela_ que participa del proyecto. Existen 5 agencias. Deberán recibir como variables de entorno los campos que representan la apuesta de una persona: nombre, apellido, DNI, nacimiento, numero apostado (en adelante 'número'). Ej.: `NOMBRE=Santiago Lionel`, `APELLIDO=Lorca`, `DOCUMENTO=30904465`, `NACIMIENTO=1999-03-17` y `NUMERO=7574` respectivamente.

Los campos deben enviarse al servidor para dejar registro de la apuesta. Al recibir la confirmación del servidor se debe imprimir por log: `action: apuesta_enviada | result: success | dni: ${DNI} | numero: ${NUMERO}`.



#### Servidor
Emulará a la _central de Lotería Nacional_. Deberá recibir los campos de la cada apuesta desde los clientes y almacenar la información mediante la función `store_bet(...)` para control futuro de ganadores. La función `store_bet(...)` es provista por la cátedra y no podrá ser modificada por el alumno.
Al persistir se debe imprimir por log: `action: apuesta_almacenada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

#### Comunicación:
Se deberá implementar un módulo de comunicación entre el cliente y el servidor donde se maneje el envío y la recepción de los paquetes, el cual se espera que contemple:
* Definición de un protocolo para el envío de los mensajes.
* Serialización de los datos.
* Correcta separación de responsabilidades entre modelo de dominio y capa de comunicación.
* Correcto empleo de sockets, incluyendo manejo de errores y evitando los fenómenos conocidos como [_short read y short write_](https://cs61.seas.harvard.edu/site/2018/FileDescriptors/).


### Ejercicio N°6:
Modificar los clientes para que envíen varias apuestas a la vez (modalidad conocida como procesamiento por _chunks_ o _batchs_). 
Los _batchs_ permiten que el cliente registre varias apuestas en una misma consulta, acortando tiempos de transmisión y procesamiento.

La información de cada agencia será simulada por la ingesta de su archivo numerado correspondiente, provisto por la cátedra dentro de `.data/datasets.zip`.
Los archivos deberán ser inyectados en los containers correspondientes y persistido por fuera de la imagen (hint: `docker volumes`), manteniendo la convencion de que el cliente N utilizara el archivo de apuestas `.data/agency-{N}.csv` .

En el servidor, si todas las apuestas del *batch* fueron procesadas correctamente, imprimir por log: `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`. En caso de detectar un error con alguna de las apuestas, debe responder con un código de error a elección e imprimir: `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`.

La cantidad máxima de apuestas dentro de cada _batch_ debe ser configurable desde config.yaml. Respetar la clave `batch: maxAmount`, pero modificar el valor por defecto de modo tal que los paquetes no excedan los 8kB. 

Por su parte, el servidor deberá responder con éxito solamente si todas las apuestas del _batch_ fueron procesadas correctamente.

### Ejercicio N°7:

Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo.
Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia.
Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.

El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`.
Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo no se podrán responder consultas por la lista de ganadores con información parcial.

Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.

No es correcto realizar un broadcast de todos los ganadores hacia todas las agencias, se espera que se informen los DNIs ganadores que correspondan a cada una de ellas.

## Parte 3: Repaso de Concurrencia
En este ejercicio es importante considerar los mecanismos de sincronización a utilizar para el correcto funcionamiento de la persistencia.

### Ejercicio N°8:

Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo. En caso de que el alumno implemente el servidor en Python utilizando _multithreading_,  deberán tenerse en cuenta las [limitaciones propias del lenguaje](https://wiki.python.org/moin/GlobalInterpreterLock).

## Condiciones de Entrega
Se espera que los alumnos realicen un _fork_ del presente repositorio para el desarrollo de los ejercicios y que aprovechen el esqueleto provisto tanto (o tan poco) como consideren necesario.

Cada ejercicio deberá resolverse en una rama independiente con nombres siguiendo el formato `ej${Nro de ejercicio}`. Se permite agregar commits en cualquier órden, así como crear una rama a partir de otra, pero al momento de la entrega deberán existir 8 ramas llamadas: ej1, ej2, ..., ej7, ej8.
 (hint: verificar listado de ramas y últimos commits con `git ls-remote`)

Se espera que se redacte una sección del README en donde se indique cómo ejecutar cada ejercicio y se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado (Parte 2) y los mecanismos de sincronización utilizados (Parte 3).

Se proveen [pruebas automáticas](https://github.com/7574-sistemas-distribuidos/tp0-tests) de caja negra. Se exige que la resolución de los ejercicios pase tales pruebas, o en su defecto que las discrepancias sean justificadas y discutidas con los docentes antes del día de la entrega. El incumplimiento de las pruebas es condición de desaprobación, pero su cumplimiento no es suficiente para la aprobación. Respetar las entradas de log planteadas en los ejercicios, pues son las que se chequean en cada uno de los tests.

La corrección personal tendrá en cuenta la calidad del código entregado y casos de error posibles, se manifiesten o no durante la ejecución del trabajo práctico. Se pide a los alumnos leer atentamente y **tener en cuenta** los criterios de corrección informados  [en el campus](https://campusgrado.fi.uba.ar/mod/page/view.php?id=73393).

## Resolución de Ejercicios

### Ejercicio 1

Para resolver el ejercicio 1 se utilizó un script de python `mi-generador.py` para llamarlo desde generar-compose.sh. Inicialmente este fue implementado con líneas hardcodeadas según el archivo ya presente `docker-compose-dev.yaml`, pero a medida que fue evolucionando el TP (y por lógica común) se vio la necesidad de realizar un procesamiento más dinámico; por lo que se decidió desde `mi-generador.py` leer dicho archivo identificando las secciones para generar el archivo.
  
#### Ejecución

Para ejecutar el generador de archivos se debe utilizar el siguiente comando:

```bash
./generar-compose.sh <nombre_archivo_salida> <numero_clientes>
```

### Ejercicio 2

Para hacer persistir los archivos de configuración fuera de la imagen se utilizaron volúmenes:

    volumes:
      - ./server/config.ini:/config.ini

    volumes:
      - ./client/config.yaml:/config.yaml

Para su funcionamiento correcto, también se tuvo que eliminar algunas variables de entorno definidas en el archivo `docker-compose-dev.yaml`.
  
#### Ejecución

Tras modificar los archivos de config, ejecutar los siguientes comandos:

```bash
./generar-compose.sh <nombre_archivo_salida> <numero_clientes>
make docker-compose-up 
```

### Ejercicio 3

Para la resolución del ejercicio 3 se creó el script `validar-echo-server.sh`, donde se levanta un container provisorio en la misma red que el servidor, para luego enviarle un mensaje y validar que la respuesta sea la esperada. En particular, se utilizó el comando `echo` para hacer una simple comparación entre el mensaje enviado y el recibido.

#### Ejecución

Para validar la implementación, se debe ejecutar el siguiente comando tras levantar los contenedores:

```bash
./validar-echo-server.sh
```

### Ejercicio 4

Para manejar la señal SIGTERM de forma graceful se modificó tanto el cliente como el servidor. Esto implica que ambas aplicaciones capturan la señal SIGTERM usando signal handlers, cierran todos los recursos y loguean el cierre de cada recurso.

#### Servidor (Python):
- Se agregó un signal handler para SIGTERM que establece una bandera `_running = False`
- El loop principal del servidor verifica esta bandera para terminar el bucle
- Se cierra el socket del servidor en el handler de señales y en el bloque finally
- Se logea cada cierre de conexión cliente

#### Cliente (Go):
- Se utiliza un channel para capturar la signal SIGTERM
- El loop del cliente se ejecuta en una goroutine separada
- Se implementó un método `Shutdown()` que establece una bandera de cierre y cierra conexiones
- El main function espera tanto la terminación normal como las señales para actuar apropiadamente

#### Ejecución

Para probar la salida graceful se puede levantar los contenedores y luego enviar una señal SIGTERM:

```bash
make docker-compose-up

docker kill -s SIGTERM <container_id>
```

Luego para comprobar que el cierre fue graceful, ver los logs de los contenedores:

```
docker logs <container_id>
```

### Ejercicio 5

#### Protocolo de Comunicación

Para evitar short reads/writes, se implementó un protocolo binario que envía primero la longitud total del mensaje para que el receptor pueda leer exactamente la cantidad de bytes necesarios de una vez. Esto elimina la necesidad de múltiples llamadas `read()` y mejora la eficiencia de la comunicación. Para minimizar la sobrecarga de datos, todos los datos que pueden ser pasados a números enteros se envían parseados como tales.

**Mensaje de Apuesta (Cliente a Servidor):**
```
| Campo        | Tipo      | Tamaño    | Descripción                    |
|--------------|-----------|-----------|--------------------------------|
| message_len  | uint32    | 4 bytes   | Longitud total del mensaje     |
| client_id    | uint32    | 4 bytes   | ID del cliente (big endian)    |
| nombre_len   | uint32    | 4 bytes   | Longitud del nombre            |
| nombre       | string    | variable  | Nombre del apostador           |
| apellido_len | uint32    | 4 bytes   | Longitud del apellido          |
| apellido     | string    | variable  | Apellido del apostador         |
| documento    | uint32    | 4 bytes   | DNI del apostador              |
| nacimiento   | uint32    | 4 bytes   | Fecha YYYYMMDD                 |
| numero       | uint32    | 4 bytes   | Número apostado                |
```

**Respuesta (Servidor a Cliente):**
```
| Campo     | Tipo   | Tamaño  | Descripción                |
|-----------|--------|---------|----------------------------|
| response  | uint8  | 1 byte  | 0=OK, 1=ERROR              |
```

Para separar la capa de dominio y la capa de comunicación, todo lo relacionado al protocolo se encuentra en archivos `protocol.go` y `protocol.py`.

#### Manejo de Errores

- Se manejan errores de conexión, serialización/deserialización y validación
- Logging detallado de errores con contexto específico

#### Configuración

Los clientes reciben datos de apuesta a través de variables de entorno:
- `CLI_NOMBRE`: Nombre del apostador
- `CLI_APELLIDO`: Apellido del apostador  
- `CLI_DOCUMENTO`: DNI del apostador
- `CLI_NACIMIENTO`: Fecha de nacimiento (YYYY-MM-DD)
- `CLI_NUMERO`: Número apostado

#### Ejecución

Para el ejercicio 5 se incorporaron datos de prueba en el archivo `docker-compose-dev.yaml` y en el script `mi-generador.py` (para hasta 5 clientes).

Si se desea ejecutar el sistema y probar el flujo completo:

```bash
./generar-compose.sh docker-compose-dev.yaml 5

make docker-compose-up

make docker-compose-logs
```

### Ejercicio 6

#### Modificaciones del Protocolo

Se extendió el protocolo binario para soportar batches de apuestas:

**Mensaje de Batch (Cliente a Servidor):**
```
| Campo         | Tipo      | Tamaño    | Descripción                    |
|---------------|-----------|-----------|--------------------------------|
| message_len   | uint32    | 4 bytes   | Longitud total del mensaje     |
| client_id     | uint32    | 4 bytes   | ID del cliente (big endian)    |
| num_apuestas  | uint32    | 4 bytes   | Cantidad de apuestas en batch  |
| apuestas      | apuesta[] | variable  | Array de apuestas              |
```

Donde el tipo apuesta hace referencia a la estructura de datos que representa una única apuesta. Cada apuesta dentro del batch mantiene la misma estructura del ejercicio 5 pero sin client_id.

#### Batch: maxAmount

El valor del batch: maxAmount fue configurado tras un análisis de los archivos .csv proporcionados por la cátedra; se verificó que la combinación de Nombre y Apellido más larga era "Milagros De Los Angeles" + "Valenzuela", lo que resultaría en 33 bytes, constituyendo una apuesta de longitud total 53 bytes (sumado a los 20 bytes fijos). Dado esto, para que no exceda el tamaño máximo de 8kB, se estableció un límite conservador de 150 apuestas por batch: $8+53*150=7958$. Para evitar cargar en memoria todo el archivo, se declara una variable `currentBatch` que se vacía después de enviar cada batch.

Para poder acceder a los mencionados archivos .csv, se agregó un volume en el archivo `docker-compose-dev.yaml`.

#### Ejecución

Para comprobar el funcionamiento del sistema y el procesamiento de batches:

```bash
./generar-compose.sh docker-compose-dev.yaml 5

make docker-compose-up

make docker-compose-logs
```

### Ejercicio 7

#### Modificaciones del Protocolo

Se extendió el protocolo para soportar diferentes tipos de mensajes mediante la inclusión de un campo `message_type` al inicio de cada mensaje:

**Tipos de Mensajes:**
- `MessageTypeBatch = 1`: Envío de lote de apuestas
- `MessageTypeFinishedSending = 2`: Notificación de fin de envío de apuestas  
- `MessageTypeQueryWinners = 3`: Consulta de ganadores

**Mensaje de Notificación de Fin (Cliente a Servidor):**
```
| Campo        | Tipo      | Tamaño    | Descripción                    |
|--------------|-----------|-----------|--------------------------------|
| message_len  | uint32    | 4 bytes   | Longitud total del mensaje     |
| message_type | uint32    | 4 bytes   | Tipo de mensaje (2)            |
| client_id    | uint32    | 4 bytes   | ID del cliente (big endian)    |
```

**Mensaje de Consulta de Ganadores (Cliente a Servidor):**
```
| Campo        | Tipo      | Tamaño    | Descripción                    |
|--------------|-----------|-----------|--------------------------------|
| message_len  | uint32    | 4 bytes   | Longitud total del mensaje     |
| message_type | uint32    | 4 bytes   | Tipo de mensaje (3)            |
| client_id    | uint32    | 4 bytes   | ID del cliente (big endian)    |
```

**Respuesta de Ganadores (Servidor a Cliente):**
```
| Campo         | Tipo      | Tamaño    | Descripción                    |
|---------------|-----------|-----------|--------------------------------|
| response      | uint8     | 1 byte    | 0=OK, 1=ERROR                  |
| num_ganadores | uint32    | 4 bytes   | Cantidad de ganadores          |
| ganadores     | uint32[]  | variable  | Array de DNIs ganadores        |
```

#### Flujo del Sistema

1. **Envío de Apuestas**: Los clientes procesan y envían todas las apuestas por batches
2. **Notificación de Fin**: Cada cliente envía una notificación cuando termina de enviar todas sus apuestas
3. **Sorteo**: El servidor espera las notificaciones de las 5 agencias y luego realiza el sorteo
4. **Consulta de Ganadores**: Los clientes consultan inmediatamente por sus ganadores específicos, y en caso de obtener error(que indica sorteo no realizado), esperan 1 segundo y reintentan
5. **Respuesta de Ganadores**: El servidor responde con los DNIs ganadores de cada agencia

El servidor mantiene:
- `_finished_agencies`: Set que registra las agencias que terminaron de enviar apuestas
- `_lottery_done`: Flag que indica si el sorteo ya fue realizado

Cuando todas las agencias del sistema notifican que terminaron, el servidor logea: `action: sorteo | result: success`.

Las consultas de ganadores antes del sorteo son rechazadas con error, y los clientes reintentan cada segundo para obtener los resultados.

#### Ejecución

Para probar el flujo completo del sorteo:

```bash
./generar-compose.sh docker-compose-dev.yaml 5

make docker-compose-up

make docker-compose-logs
```

### Ejercicio 8

Se modificó el servidor para manejar conexiones y procesar mensajes en paralelo utilizando multithreading; el servidor principal acepta conexiones en un loop continuo, y cada nueva conexión cliente se maneja en un thread separado. A pesar de las limitaciones del GIL de Python, esto sigue siendo beneficioso al tratarse de operaciones I/O.

#### Mecanismos de Sincronización

Dado que las funciones `store_bets()` y `load_bets()` no son thread-safe, se implementaron tres locks principales:

1. **`_storage_lock`**: Protege las operaciones de almacenamiento (`store_bets` y `load_bets`)
   - Utilizado en `__handle_bet_batch()` para escritura segura de apuestas
   - Utilizado in `__get_winners_for_agency()` para lectura segura de apuestas

2. **`_finished_agencies_lock`**: Protege el set de agencias que terminaron
   - Evita condiciones de carrera al agregar agencias al set
   - Garantiza consistencia al verificar el conteo de agencias terminadas

3. **`_lottery_lock`**: Protege las operaciones relacionadas con el estado del sorteo
   - Asegura que solo un thread pueda marcar el sorteo como realizado
   - Protege las consultas del estado `_lottery_done`

#### Ejecución

Para probar la funcionalidad con concurrencia:

```bash
./generar-compose.sh docker-compose-dev.yaml 5

make docker-compose-up

make docker-compose-logs
```

