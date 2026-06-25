/** biome-ignore-all lint/a11y/useMediaCaption: 233 */
import { format } from "date-fns";
import { Languages, Loader2 } from "lucide-react";
import { useState } from "react";
import useChatStore from "@/stores/chat-store";
import type { ChatRoom, Message, UserProfile } from "../../types/chat";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { Dialog, DialogContent } from "../ui/dialog";

interface MessageBubbleProps {
  message: Message;
  room: ChatRoom;
  userProfile: UserProfile | null;
}

const MessageBubble = ({ message, room, userProfile }: MessageBubbleProps) => {
  const [isMediaDialogOpen, setIsMediaDialogOpen] = useState(false);
  const translation = useChatStore(
    (state) => state.translationByMessageId[message.id],
  );
  const toggleTranslation = useChatStore((state) => state.toggleTranslation);

  // 判断是否是自己发的消息：发送者ID 不等于 聊天室对方的 userId
  const isMine = message.senderName !== room.displayName;

  const formatTime = (date: string) => {
    return format(new Date(date), "yyyy-MM-dd HH:mm:ss");
  };

  const handleMediaClick = () => {
    setIsMediaDialogOpen(true);
  };

  const renderMediaContent = () => {
    if (!message.mediaUrl) return null;

    const commonProps = {
      className:
        "max-h-[160px] rounded-md cursor-pointer transition-opacity hover:opacity-80",
      onClick: handleMediaClick,
    };

    if (message.type === "image") {
      return <img src={message.mediaUrl} alt="聊天图片" {...commonProps} />;
    }

    if (message.type === "video") {
      return (
        <video
          src={message.mediaUrl}
          controls={false}
          muted
          preload="metadata"
          {...commonProps}
        />
      );
    }

    return null;
  };

  const renderFullscreenMedia = () => {
    if (!message.mediaUrl) return null;

    if (message.type === "image") {
      return (
        <img
          src={message.mediaUrl}
          alt="聊天图片"
          className="max-w-full max-h-full object-contain rounded-sm"
        />
      );
    }

    if (message.type === "video") {
      return (
        <video
          src={message.mediaUrl}
          controls
          autoPlay
          className="max-w-full max-h-full"
        />
      );
    }

    return null;
  };

  return (
    <div
      className={`flex items-start mb-4 gap-3 ${isMine ? "flex-row-reverse" : ""}`}
    >
      {/* 对方头像 */}
      {!isMine && (
        <Avatar className="w-8 h-8 md:w-10 md:h-10 shrink-0">
          <AvatarImage src={room.avatarUrl} alt={room.displayName} />
          <AvatarFallback>{room.displayName.slice(0, 1) || "R"}</AvatarFallback>
        </Avatar>
      )}
      {/* 自己的头像 */}
      {isMine && (
        <Avatar className="w-8 h-8 md:w-10 md:h-10 shrink-0">
          <AvatarImage src={userProfile?.avatarUrl} alt={userProfile?.displayName} />
          <AvatarFallback className="text-xs">我</AvatarFallback>
        </Avatar>
      )}

      <div
        className={`flex flex-col max-w-[70%] ${isMine ? "items-end" : "items-start"}`}
      >
        {/* 名字 + 时间 */}
        <div
          className={`flex items-center gap-2 mb-1 ${isMine ? "flex-row-reverse" : ""}`}
        >
          <span className="text-sm font-medium text-foreground">
            {isMine ? message.senderName : room.displayName}
          </span>
          <span className="text-xs text-muted-foreground/70">
            {formatTime(message.createdAt)}
          </span>
          {message.type === "text" && message.content.trim() && (
            <button
              type="button"
              onClick={() => void toggleTranslation(message)}
              className="text-muted-foreground/70 hover:text-foreground transition-colors"
              title={translation?.visible ? "隐藏翻译" : "翻译"}
            >
              {translation?.loading ? (
                <Loader2 className="h-3 w-3 animate-spin" />
              ) : (
                <Languages className="h-3 w-3" />
              )}
            </button>
          )}
        </div>

        {/* 气泡：自己的消息用主题色显示在右边，对方的用默认卡片色显示在左边 */}
        <div
          className={`rounded-lg ${
            isMine
              ? "bg-primary text-primary-foreground"
              : "bg-card"
          }`}
        >
          {message.type === "text" && (
            <div className="px-3 py-2">
              <p className="text-sm whitespace-pre-wrap break-words">
                {message.content}
              </p>
              <div
                className={`transition-all duration-300 ease-in-out overflow-hidden ${
                  translation?.visible
                    ? "max-h-96 mt-2 border-t border-border/20"
                    : "max-h-0"
                }`}
              >
                <div className="w-full my-2 h-[1px] bg-muted-foreground/20" />
                <div className="text-xs text-muted-foreground/70 mb-1">
                  翻译来自谷歌机翻，仅供参考：
                </div>
                {translation?.loading && (
                  <p className="text-sm text-muted-foreground">翻译中...</p>
                )}
                {translation?.error && (
                  <p className="text-sm text-destructive">
                    {translation.error}
                  </p>
                )}
                {translation?.text && (
                  <p className="text-sm text-muted-foreground whitespace-pre-wrap break-words">
                    {translation.text}
                  </p>
                )}
              </div>
            </div>
          )}

          {(message.type === "image" || message.type === "video") && (
            <div className="space-y-2 p-1">
              {renderMediaContent()}
              {message.content !==
                `[${message.type === "image" ? "图片" : "视频"}]` && (
                <p className="text-sm text-muted-foreground px-2 pb-1">
                  {message.content}
                </p>
              )}
            </div>
          )}
        </div>
      </div>

      {message.mediaUrl && (
        <Dialog open={isMediaDialogOpen} onOpenChange={setIsMediaDialogOpen}>
          <DialogContent className="p-0 bg-transparent border-none">
            <div className="flex items-center justify-center w-full">
              {renderFullscreenMedia()}
            </div>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
};

export default MessageBubble;
