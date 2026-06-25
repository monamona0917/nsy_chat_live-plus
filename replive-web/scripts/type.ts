export interface Schema {
  field1?: Message[];
  field2?: string;
  field3?: string;
}

export interface Message {
  field1?: string; // id
  field2?: string;
  field3?: string;
  field4?: UserInfo;
  field5?: string; // message type (1: text, 2: image, 3: video)
  field6?: string; // text
  field7?: string; // message image url
  field8?: string; // message video url
  field9?: string; // message video thumbnail url
  field100: field1_typefield100_type;
}

export interface UserInfo {
  field1?: string; // id
  field2?: string; // username
  field3?: string; // displayName
  field4?: string; // avatar
  field5?: string;
  field7?: string;
  field12?: string; // bio
  field13?: string; // background image url
  field14?: string;
  field15?: string; // icon
  field16?: field4_typefield16_type;
  field17?: string;
  field18?: string;
  field19?: string;
  field20?: string;
  field21?: string; // fan name
  field22?: string;
  field23?: string;
}

export interface field1_typefield100_type {
  field1: string; // createdAt
  field2?: string;
}

export interface field4_typefield16_type {
  field1: string; // createdAt
  field2?: string;
}

export interface MessageData {
  // id: string;
  content: string;
  type: "text" | "image" | "video";
  createdAt: string;
  mediaUrl?: string;
  // user: {
  //   username: string;
  //   nickname: string;
  //   avatar: string;
  //   createdAt: string;
  // };
}
