package inc.asapp.flink.apps

import java.util.Properties
import java.util.concurrent._

import org.apache.flink.streaming.api.scala._
import org.apache.flink.streaming.connectors.kafka.FlinkKafkaConsumer010
import org.apache.flink.streaming.util.serialization.SimpleStringSchema
import org.apache.flink.streaming.util.serialization.{ DeserializationSchema, SerializationSchema }

import org.apache.flink.api.common.restartstrategy._
import org.apache.flink.api.common.time.Time
import org.apache.flink.streaming.api.CheckpointingMode
import org.apache.flink.streaming.api.environment.CheckpointConfig.ExternalizedCheckpointCleanup

import scala.concurrent.duration._
import com.typesafe.config.ConfigFactory

object Archiver {
  def main(args: Array[String]): Unit = {
    val appname = "Archiver-PersistKafka"

    // set up the execution environment
    val env = StreamExecutionEnvironment.getExecutionEnvironment

    setupCheckpointing(env)

    val properties = new Properties()
    properties.setProperty("bootstrap.servers", "kafka:9092")
    properties.setProperty("group.id", appname)

    val stream: DataStream[String] = env.addSource(new FlinkKafkaConsumer010[String](
      "first_topic",
      new SimpleStringSchema(),
      properties
    ))

    stream.print

    env.execute(s"started the flink app - $appname")
  }
  
  def setupCheckpointing(env: StreamExecutionEnvironment): Unit = {
    // TODO: values right now are hardcoded but need to be pulled out of the config file
    env.enableCheckpointing(60*1000)
    env.getCheckpointConfig.setCheckpointingMode(CheckpointingMode.EXACTLY_ONCE)
    env.getCheckpointConfig.setCheckpointTimeout(30*1000)
    env.getCheckpointConfig.setMaxConcurrentCheckpoints(1)
    env.getCheckpointConfig.enableExternalizedCheckpoints(ExternalizedCheckpointCleanup.RETAIN_ON_CANCELLATION)
    env.getCheckpointConfig.setFailOnCheckpointingErrors(false)
    env.setRestartStrategy(RestartStrategies.fixedDelayRestart(30, Time.of(1, TimeUnit.MINUTES)))
  }
}
