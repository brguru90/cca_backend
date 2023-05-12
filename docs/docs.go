// Code generated by swaggo/swag. DO NOT EDIT.

package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/admin/upload_streaming_video/": {
            "post": {
                "description": "api to upload video content for multiple adaptive bit rate streaming",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Video upload"
                ],
                "summary": "video upload",
                "parameters": [
                    {
                        "type": "file",
                        "description": "File",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    }
                }
            }
        },
        "/login": {
            "post": {
                "description": "allow people to login into their account",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Login"
                ],
                "summary": "url to login",
                "parameters": [
                    {
                        "description": "Add user",
                        "name": "existing_user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user_views.UserCredential"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    }
                }
            }
        },
        "/login_mobile": {
            "post": {
                "description": "allow people to login into their account",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Mobile Login"
                ],
                "summary": "url to login with mobile number",
                "parameters": [
                    {
                        "description": "Add user",
                        "name": "existing_user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user_views.UserMobileCredential"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    }
                }
            }
        },
        "/login_status": {
            "post": {
                "description": "api used to validate user login session",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Login status"
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    }
                }
            }
        },
        "/sign_up": {
            "post": {
                "description": "allow people to create new to user account",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "SignUp"
                ],
                "summary": "url to signup",
                "parameters": [
                    {
                        "description": "Add user",
                        "name": "new_user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user_views.UserCredential"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    }
                }
            }
        },
        "/user/": {
            "get": {
                "description": "allow people to view their user profile data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "View user data"
                ],
                "summary": "url to view user data",
                "parameters": [
                    {
                        "type": "string",
                        "description": "page",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "limit",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    }
                }
            },
            "put": {
                "description": "allow people to update their user profile data",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Update user data"
                ],
                "summary": "url to update user data",
                "parameters": [
                    {
                        "description": "Add user",
                        "name": "new_user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/database.UsersModel"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    }
                }
            },
            "delete": {
                "description": "allow people to delete their account",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Delete user account"
                ],
                "summary": "url to delete user account",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    }
                }
            }
        },
        "/user/active_sessions/": {
            "get": {
                "description": "return the active user session/token across all browser",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Get Active sessions"
                ],
                "summary": "get active user login session",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    }
                }
            }
        },
        "/user/block_token/": {
            "post": {
                "description": "Adds the token of user to block list based on provided token id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Block sessions"
                ],
                "summary": "block specified session",
                "parameters": [
                    {
                        "description": "block token",
                        "name": "block_active_session",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/database.ActiveSessionsModel"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    }
                }
            }
        },
        "/user/logout/": {
            "get": {
                "description": "API allow user to logout, which delete the cookie which stores token",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Logout"
                ],
                "summary": "allow user to logout",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    }
                }
            }
        },
        "/verify_social_auth": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "VerifySocialAuth"
                ],
                "summary": "url to signup/login with social authentication",
                "parameters": [
                    {
                        "description": "Add user",
                        "name": "new_or_existing_user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user_views.SocialAuth"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/my_modules.ResponseFormat"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "database.ActiveSessionsModel": {
            "type": "object",
            "properties": {
                "_id": {
                    "type": "string"
                },
                "createdAt": {
                    "type": "string"
                },
                "exp": {
                    "type": "integer"
                },
                "ip": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                },
                "token_id": {
                    "type": "string"
                },
                "ua": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "database.UsersModel": {
            "type": "object",
            "required": [
                "access_level",
                "email",
                "mobile",
                "password",
                "username"
            ],
            "properties": {
                "access_level": {
                    "type": "string"
                },
                "auth_provider": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "mobile": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "uid": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "my_modules.ResponseFormat": {
            "type": "object",
            "required": [
                "data",
                "msg",
                "status"
            ],
            "properties": {
                "data": {},
                "msg": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "user_views.SocialAuth": {
            "type": "object",
            "required": [
                "idToken"
            ],
            "properties": {
                "idToken": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "user_views.UserCredential": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "user_views.UserMobileCredential": {
            "type": "object",
            "required": [
                "mobile",
                "password"
            ],
            "properties": {
                "mobile": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
