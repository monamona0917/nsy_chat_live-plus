export interface User {
  id: string;
  name: string;
  avatar?: string;
}

export interface ChatRoom {
  id?: number;
  userId: string;
  chatRoomId: string;
  displayName: string;
  avatarUrl?: string;
}

export interface Message {
  id: string;
  backendId: number;
  chatMessageId: string;
  content: string;
  type: "text" | "image" | "video";
  createdAt: string;
  mediaUrl?: string;
  senderId: string;   // 发送者ID
  senderName: string; // 发送者名字
}

export interface MessageGroup {
  date: string;
  messages: Message[];
}

export interface UserProfile {
  userId: string;
  uniqueId: string;
  displayName: string;
  avatarUrl?: string;
  sendChatEnabled?: boolean;
}

export interface ChatStats {
  totalMessages: number;
  messagesByDate: Record<string, number>;
  mediaCount: {
    images: number;
    videos: number;
  };
  firstMessageDate: string;
  lastMessageDate: string;
}

export interface TranslationState {
  text?: string;
  loading?: boolean;
  error?: string;
  visible?: boolean;
}

export interface BackendChatRoom {
  id?: number;
  user_id: string;
  chat_room_id: string;
  display_name: string;
  avatar_url?: string;
}

export interface BackendUserProfile {
  user_id: string;
  unique_id: string;
  display_name: string;
  avatar_url?: string;
  send_chat?: boolean;
}

export interface BackendMessage {
  id: number;
  user_id: string;
  display_name: string;
  chat_room_id: string;
  chat_message_id: string;
  msg_type: number;
  content: string;
  image_url?: string;
  video_url?: string;
  send_time?: number;
  time_str?: string;
}

export interface ChatMessagesPage {
  messages: Message[];
  nextCursorId: number;
  prevCursorId: number;
  hasMore: boolean;
  hasOlder: boolean;
  hasNewer: boolean;
  anchorId: number;
}
