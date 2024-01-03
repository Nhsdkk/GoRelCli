# GoRelCli

###### Cli for better golang object-relational mapper inspired by prisma and GORM.

[![Publish app](https://github.com/Nhsdkk/GoRelCli/actions/workflows/publish.yml/badge.svg?branch=master)](https://github.com/Nhsdkk/GoRelCli/actions/workflows/publish.yml)

### Table of contents

---
1. [What does it do?](#what-does-it-do)
2. [How to specify schema?](#how-to-specify-schema)
3. [Relations](#relations)
4. [How to run migrations](#how-to-run-migrations)
5. [How to run generator](#how-to-run-generator)
6. [How to run clean](#how-to-run-clean)

### What does it do?

---
* Creates migrations using schema specified in YAML file
* Generates structs using specified schema
* Cleans up the names inside schema to match the rules 

### How to specify schema?

---
You can specify schema in YAML files. Here is an example of the schema file:
```yaml
connection:
  provider: postgresql
  url: env("DATABASE_URL")

models:
  - name: User
    properties:
      - name: id
        type: int
        default: autoincrement()
        id: true
      - name: email
        type: string
        unique: true
      - name: username
        type: string?
      - name: isVerified
        type: boolean
        default: false
      - name: userType
        type: UserRole
      - name: todos
        type: Todo[]
      - name: videos
        type: UserToVideoRelation[]

  - name: Todo
    properties:
      - name: id
        type: string
        default: uuid()
        id: true
      - name: title
        type: string
      - name: userId
        type: int
      - name: user
        type: User
        relationField: userId
        referenceField: id
      - name: note
        type: Note

  - name: Note
    properties:
      - name: id
        type: string
        default: uuid()
        id: true
      - name: text
        type: string
      - name: todo
        type: Todo
        relationField: id
        referenceField: id

  - name: UserToVideoRelation
    properties:
      - name: id
        type: int
        default: autoincrement()
        id: true
      - name: userId
        type: int
      - name: videoId
        type: int
      - name: user
        type: User
        relationField: userId
        referenceField: id
      - name: video
        type: Video
        relationField: videoId
        referenceField: id

  - name: Video
    properties:
      - name: id
        type: int
        default: autoincrement()
        id: true
      - name: title
        type: string
      - name: users
        type: UserToVideoRelation[]
enums:
  - name: UserRole
    values:
      - Admin
      - User
```

Let's dive into some details:
* #### Connection
  * ##### Purpose
    * Here you can specify your db provider as well as connection string (url). postgresql is the only available option for now.
    * Instead of specifying url explicitly you can use env("YOUR_ENV_VARIABLE_NAME") function to load env variable and use it as url.
  * ##### Requirements
    * Should have _**name**_ and _**url**_ property
* #### Models
  * ##### Purpose
    * Here you can specify models and properties that will correspond to them. For each of these models, new table will be created with name you specified in _**'name'**_ option.
  * ##### Requirements
    * Should have _**name**_ and _**properties**_ property
    * Should have 2 or more properties and one of them should have _**id**_ property set to _**true**_
* #### Enums
  * ##### Purpose
    * Here you can specify enums with corresponding values, that will be created.
  * ##### Requirements
    * Should have _**name**_ and _**values**_ property with 2 or more string values
* #### Properties (inside model)
  * ##### Purpose
    * Here you can specify properties on model (columns in db)
  * ##### Requirements
    * Should have _**name**_ and _**type**_ properties.
    * <details><summary>Type property should have one of these values</summary> <ul><li>int</li><li>boolean</li><li>float</li><li>string</li><li>dateTime</li><li>Models defined in schema</li><li>Enums defined in schema</li><li>Arrays (T[])</li><li>Nullable types (T?)</li></ul></details>
  * ##### Optional fields
    * id
      * Defines if field is an id field or not (id field == primary key field)
    * Default
      * Defines default value which will be assigned to cell, when row will be created
      * <details><summary>Possible values</summary> <ul><li>int</li><li>boolean</li><li>float</li><li>string</li><li>dateTime</li><li>Enums defined in schema</li><li>now() function</li><li>uuid() function</li><li>autoincrement() function</li></ul></details>
  
### Relations

---
* #### Basic Info
  * ##### Defining relations
    * Relations can be defined using _**relationField**_ and _**referenceField**_ properties
    * _**relationField**_ should be a field on model that you are defining relation on
    * _**referenceField**_ should be a field on model that you are referencing
    * The referenced model should have a field with the type of the model that you are defining relation on
* #### Supported relation types
  - [x] One to many
    * ##### Requirements
      * Should have _**relationField**_ and _**referenceField**_ properties
      * _**relationField**_ should be an array type
      * _**referenceField**_ should be a unique type
  - [x] One to one
    * ##### Requirements
      * Should have _**relationField**_ and _**referenceField**_ properties
      * _**relationField**_ should be a unique type
      * _**referenceField**_ should be a unique type
  - [x] Many to many
    * ##### Requirements
      * Separate linking model should be created, where you define 2 fields, that will reference models that you want to link
      * Both fields should have _**relationField**_ and _**referenceField**_ properties
      * Both fields should be an array type
      * Linking table should have _**id**_ field
      * Example
        ```yaml
        models:
          - name: User
            properties:
          - name: id
            type: int
            default: autoincrement()
            id: true
          - name: email
            type: string
            unique: true
          - name: username
            type: string?
          - name: isVerified
            type: boolean
            default: false
          - name: videos
            type: UserToVideoRelation[]
    
          - name: UserToVideoRelation
            properties:
              - name: id
                type: int
                default: autoincrement()
                id: true
              - name: userId
                type: int
              - name: videoId
                type: int
              - name: user
                type: User
                relationField: userId
                referenceField: id
              - name: video
                type: Video
                relationField: videoId
                referenceField: id
  
          - name: Video
            properties:
              - name: id
                type: int
                default: autoincrement()
                id: true
              - name: title
                type: string
              - name: users
                type: UserToVideoRelation[]
         ```
      
  
### How to run migrations

---
1. Download latest executable from GitHub
2. Run command in command line
  ```bash
  ./FOLDER_WHERE_EXECUTABLE_EXISTS/GoRelCli.exe migrate --path="./FOLDER_WHERE_GOREL_SCHEMA_EXISTS/gorel_schema.yml"
   ```

### How to run generator

---
1. Download latest executable from GitHub
2. Run command in command line
  ```bash
  ./FOLDER_WHERE_EXECUTABLE_EXISTS/GoRelCli.exe generate --path="./FOLDER_WHERE_GOREL_SCHEMA_EXISTS/gorel_schema.yml" --project_path="./ROOT_PROJECT_FOLDER"
   ```

### How to run clean

---
1. Download latest executable from GitHub
2. Run command in command line
  ```bash
  ./FOLDER_WHERE_EXECUTABLE_EXISTS/GoRelCli.exe clean --path="./FOLDER_WHERE_GOREL_SCHEMA_EXISTS/gorel_schema.yml"
   ```