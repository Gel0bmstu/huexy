## tp-go-proxy
go | jacks/pgx | openssl*

Proxy сохраняет запросы в локальной postgres базе днных.
Сертификаты генерируюся на лету с помощью openssl.

#### Умеет:
- Проксировать HTTP запросы;
- Проксировать HTTPS запросы: соединение с прокси устанавлявается с помощью метода CONNECT;
- Отправлять сохраненные (в базе) запросы как *burp repeater*;

#### Как запустить:
**Для запуска прокси сервера (хэндлит HTTP и HTTPS):**

```
git clone https://github.com/Gel0bmstu/huexy; 
cd ./huexy; 
sudo docker build -t gel0 . ;
docker run -p 5000:5000 --name gel0 -t gel0;
```

Либо: 
- создать в postgres суперюзера gel0 с паролем 1337
`CREATE USER gel0 WITH SUPERUSER PASSWORD '1337';`
- создать базу proxy
`CREATEDB -O gel0 proxy;`
- передоставить все права пользователю gel0
`GRANT ALL ON DATABASE proxy TO gel0;`

**Для запуска repeater'a:**

`
	cd ./repeater
	go run main.go
`
Чтобы получить список всех сохраненных запросов (все, что лежит в базе) необходимо выполнить запуск репитера с флагом -r:
`go run main.go -r`
После выполения команды в консоль выведится список всех сохраненных запросов из базы, каждому из запросов соответствует уникальный Id. Чтобы осуществить повторную отправку запроса с id *x*, необходимо выполнить запуск репитера с флагом -i *x*, пример:
`go run main.go -i 7`

*При каждом новом билде докера таблицы базы данных proxy отчищаются.*

#### Как пользоваться:

**Firefox:**
*Settings->Proxy->Manual proxy configuration->Set `localhost:8080` -> Use this proxy server for all protocols*
