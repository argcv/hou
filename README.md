# Hou

For the purpose of providing a web service to a Node.js project to users, we may require a small middleware to start an HTTP service to handle the HTTP requests.


## Solution: Python

In python, we can use **http.server** to start a web service.

```bash
python3 -m http.server
```

This command will return the corresponding files based on your requests.

Unfortunately, this does not work for a Node.js project. We may visit `http://example.com/foo`, but the real target is /index.html. This will raise an unexpected 404 error.

## Solution: [webpack-dev-server](https://github.com/webpack/webpack-dev-server), [webpack-serve](https://github.com/webpack-contrib/)

These toolkits provided a very useful envoronment in developing. But not designed for the production.

## Solution: Nginx

We can also configure Nginx and could get a perfect result.

The configuration file is pretty easy

```nginx
server {
    listen       80;
    server_name  example.com;

    root /path/to/the/project;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html  =404;
    }
}
```

However, the steps are kind of too much.

## Solution: Hou

Recently, I prepared a small script based on golang.  It will do the similar job as the behavior above in Nginx, but much smaller and efficacy.

### Install or update from source

If you just wish to use it. Please just use the following command:

```bash
go get -v github.com/argcv/hou
```

If you wish to make contributions to the repo. You can use the following script.

```bash
curl -L https://bit.ly/2KzJeU6 | bash
```

### Docker image

If you are using Chinese network, you may use aliyun's mirror: `registry.cn-zhangjiakou.aliyuncs.com/yuikns/hou`

Otherwise, you may use `yuikns/hou` from [dockerhub](https://hub.docker.com/r/yuikns/hou/).

Since the logic is very easy and the dependencies on very few packages, this image is [only 6MB](https://hub.docker.com/r/yuikns/hou/tags/) in the current. 

#### Example of using Docker through [docker-compose](https://docs.docker.com/compose/) 

A [docker-compose.yml](https://docs.docker.com/compose/compose-file/) file is a YAML file that defines how Docker containers should behave in production.

Here is an example to use Hou easily

```yaml
version: '3'

services:
  hou:
    image: yuikns/hou:latest
    ports:
      - 6789:6789
    restart: always
    volumes:
      # /app is the default root of the service
      - ./dist:/app 
#    command: ["-d", "-v" ] # other options
```
