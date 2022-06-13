package main

import "flag"

var (
	mongodbUsername = flag.String("username", "root", "mongoDB username")
	mongoDBpassword = flag.String("password", "password", "mongoDB password")
	mongoDBHostAddr = flag.String("hostAddr", "mongodb:27017", "mongoDB host address")
	httpBindAddr    = flag.String("httpBindAddr", "0.0.0.0:80", "http server bind address")

	nodeUrl = flag.String("nodeUrl", "http://localhost:14265/", "node url")

	basicAuthUser     = flag.String("basicAuthUser", "admin", "basic auth user")
	basicAuthPassword = flag.String("basicAuthPassword", "secret", "basic auth password")
)
