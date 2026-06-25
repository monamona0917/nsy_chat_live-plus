import * as fs from "node:fs/promises";
import path from "node:path";
import * as protobuf from "protobufjs";
import type { MessageData, Schema } from "./type";

// --- 定义输入和输出 ---
// .proto 文件路径
const PROTO_FILE_PATH = path.join(__dirname, "./data/default.proto");
// 要解码的消息类型全名 (包名.消息名)
const MESSAGE_TYPE = "myapp.Schema";
// Protobuf 二进制文件路径
const BINARY_INPUT_PATH = path.join(__dirname, "./data/data.bin");
// 解码后输出的 JSON 文件路径
const JSON_OUTPUT_PATH = path.join(__dirname, "./data/data.json");

/**
 * 将原始的 Schema 对象映射为格式化的 MessageData 数组
 * @param schema 原始的 Schema 数据
 * @returns 格式化后的 MessageData 数组
 */
function mapSchemaToMessageData(schema: Schema): MessageData[] {
  // 1. 安全地获取消息数组，如果不存在则返回空数组
  const messages = schema.field1 || [];

  // 2. 使用 Array.prototype.map 遍历并转换每个 message 对象
  return messages
    .map((message): MessageData | null => {
      // 辅助函数，用于将类型码转换为可读字符串
      const getMessageType = (
        typeCode?: string,
      ): "text" | "image" | "video" => {
        switch (typeCode) {
          case "1":
            return "text";
          case "2":
            return "image";
          case "3":
            return "video";
          default:
            return "text"; // 默认返回 'text'，保证类型安全
        }
      };

      const messageType = getMessageType(message.field5);

      // 根据消息类型确定 mediaUrl
      let mediaUrl: string | undefined;
      if (messageType === "image") {
        mediaUrl = message.field7;
      } else if (messageType === "video") {
        mediaUrl = message.field8;
      }

      const fanName = message.field4?.field21 ?? null;
      if (!fanName) {
        return null;
      }

      // 3. 构建并返回 MessageData 对象
      // 使用可选链 (?.) 和空值合并运算符 (??) 来处理可能不存在的字段
      return {
        // id: message.field1 ?? "unknown_id",
        content: message.field6 ?? "",
        type: messageType,
        createdAt: new Date(+message.field100.field1 * 1000).toISOString(),
        mediaUrl: mediaUrl,
      };
    })
    .filter((message): message is MessageData => !!message) // 过滤掉 null 值
    .sort((a, b) => {
      // 按照创建时间排序，最新的在前面
      return new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
    });
}

async function decodeProtobufToJson(
  protoPath: string,
  messageType: string,
  binaryPath: string,
  jsonPath: string,
) {
  try {
    console.log("开始解码...");
    // 1. 加载 .proto 文件
    const root = await protobuf.load(protoPath);
    console.log(`  -> .proto 文件加载成功: ${protoPath}`);

    // 2. 查找具体的消息类型
    const TargetMessage = root.lookupType(messageType);
    if (!TargetMessage) {
      throw new Error(`在 ${protoPath} 中未找到消息类型: ${messageType}`);
    }
    console.log(`  -> 消息类型查找成功: ${messageType}`);

    // 3. 读取二进制文件内容
    const buffer = await fs.readFile(binaryPath);
    console.log(`  -> 二进制文件读取成功: ${binaryPath}`);

    // 4. 解码二进制数据
    const decodedMessage = TargetMessage.decode(buffer);
    console.log("  -> 二进制数据解码完成");

    // 5. 将解码后的消息转换为纯 JavaScript 对象
    //    使用 toObject 方法可以更好地处理默认值和枚举等情况
    const object = TargetMessage.toObject(decodedMessage, {
      longs: String, // 将 64 位整数(long)转换为字符串
      enums: String, // 将枚举值转换为字符串
      bytes: String, // 将字节(bytes)转换为 base64 字符串
      defaults: true, // 如果字段未提供，则包含默认值
      arrays: true, // 确保空数组被包含
      objects: true, // 确保空对象被包含
      oneofs: true, // 包含 oneof 虚拟属性
    });

    // 6. 将对象转换为格式化的 JSON 字符串
    const jsonString = JSON.stringify(object, null, 2); // null, 2 用于美化输出
    console.log("  -> 解码对象已转换为 JSON 字符串");

    // 7. 将 JSON 字符串写入文件
    await fs.writeFile(jsonPath, jsonString, "utf-8");

    console.log(`🎉 解码成功！JSON 文件已保存至: ${jsonPath}`);

    const messageDataArray = mapSchemaToMessageData(object as Schema);
    console.log(`  -> 成功映射 ${messageDataArray.length} 条消息数据`);

    await fs.writeFile(
      jsonPath.replace(".json", "_messages.json"),
      JSON.stringify(messageDataArray, null, 2),
      "utf-8",
    );
  } catch (error) {
    console.error("❌ 解码过程中发生严重错误:", error);
  }
}

// 执行解码函数
decodeProtobufToJson(
  PROTO_FILE_PATH,
  MESSAGE_TYPE,
  BINARY_INPUT_PATH,
  JSON_OUTPUT_PATH,
);
