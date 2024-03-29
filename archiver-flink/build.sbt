ThisBuild / resolvers ++= Seq(
    "Apache Development Snapshot Repository" at "https://repository.apache.org/content/repositories/snapshots/",
    Resolver.mavenLocal
)

name := "archiver"
version := "0.1"
organization := "com.asappinc"

ThisBuild / scalaVersion := "2.11.12"

val flinkVersion = "1.8.0"
val scalatestVersion = "3.0.4"
val typesafeConfigVersion = "1.3.2"
val json4sVersion = "3.5.3"

val flinkDependencies = Seq(
  "org.apache.flink" %% "flink-scala" % flinkVersion % "provided",
  "org.apache.flink" %% "flink-streaming-scala" % flinkVersion % "provided",
  "org.apache.flink" %% "flink-connector-kafka-0.11" % flinkVersion
)

val generalDependencies = Seq(
  "org.json4s" %% "json4s-native" % json4sVersion,
  "org.scalatest" %% "scalatest" % scalatestVersion % "test",
  "com.typesafe" % "config" % typesafeConfigVersion
)

lazy val root = (project in file(".")).
  settings(
    libraryDependencies ++= flinkDependencies ++ generalDependencies
  )

assembly / mainClass := Some("com.asappinc.Archiver")

assembly / assemblyOutputPath := file(sys.env("ASSEMBLY_JAR_PATH"))

// make run command include the provided dependencies
Compile / run  := Defaults.runTask(Compile / fullClasspath,
                                   Compile / run / mainClass,
                                   Compile / run / runner
                                  ).evaluated

// stays inside the sbt console when we press "ctrl-c" while a Flink programme executes with "run" or "runMain"
Compile / run / fork := true
Global / cancelable := true

// exclude Scala library from assembly
assembly / assemblyOption  := (assembly / assemblyOption).value.copy(includeScala = false)
