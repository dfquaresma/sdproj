## Template: java11-gci

Java11-GCI template uses gradle as a build system and it is the Java 11 template that run with GCI java agent.

Gradle version: 5.3

### Structure

There are three projects which make up a single gradle build:

- model - (Library) classes for parsing request/response
- function - (Library) your function code as a developer, you will only ever see this folder
- entrypoint - (App) HTTP server for re-using the JVM between requests

### Handler

The handler is written in the `./src/main/Handler.java` folder

Tests are supported with junit via files in `./src/test`

### External dependencies

External dependencies can be specified in ./build.gradle in the normal way using jcenter, a local JAR or some other remote repository.

