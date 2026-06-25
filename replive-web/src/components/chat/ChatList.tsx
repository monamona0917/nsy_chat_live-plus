import { Loader2 } from "lucide-react";
import type React from "react";
import { useEffect, useRef } from "react";
import useChatStore from "@/stores/chat-store";
import type { ChatRoom, UserProfile } from "@/types/chat";
import DateBadge from "./DateBadge";
import MessageBubble from "./MessageBubble";

const TOP_LOAD_THRESHOLD = 48;

interface ChatListProps {
  userProfile: UserProfile | null;
}

const ChatList = ({ userProfile }: ChatListProps) => {
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const restoredRoomRef = useRef<string | null>(null);
  const isFetchingOlderRef = useRef(false);

  const selectedRoom = useChatStore((state) => state.selectedRoom);
  const messageGroups = useChatStore((state) => state.messageGroups);
  const searchQuery = useChatStore((state) => state.searchQuery);
  const searchResults = useChatStore((state) => state.searchResults);
  const hasMoreByRoom = useChatStore((state) => state.hasMoreByRoom);
  const hasNewerByRoom = useChatStore((state) => state.hasNewerByRoom);
  const isLoadingMessages = useChatStore((state) => state.isLoadingMessages);
  const isLoadingMore = useChatStore((state) => state.isLoadingMore);
  const isLoadingNewer = useChatStore((state) => state.isLoadingNewer);
  const jumpTargetMessageId = useChatStore((state) => state.jumpTargetMessageId);
  const scrollToBottomToken = useChatStore((state) => state.scrollToBottomToken);
  const loadOlderMessages = useChatStore((state) => state.loadOlderMessages);
  const loadNewerMessages = useChatStore((state) => state.loadNewerMessages);
  const jumpToMessage = useChatStore((state) => state.jumpToMessage);
  const clearJumpTarget = useChatStore((state) => state.clearJumpTarget);

  const selectedRoomKey = selectedRoom ? roomKey(selectedRoom) : "";
  const isSearchMode = searchQuery.length > 0;
  const hasMoreHistory =
    !isSearchMode && selectedRoom
      ? (hasMoreByRoom[selectedRoomKey] ?? false)
      : false;
  const hasNewerMessages =
    !isSearchMode && selectedRoom
      ? (hasNewerByRoom[selectedRoomKey] ?? false)
      : false;

  useEffect(() => {
    if (restoredRoomRef.current !== selectedRoomKey) {
      restoredRoomRef.current = null;
    }
  }, [selectedRoomKey]);

  useEffect(() => {
    const container = scrollContainerRef.current;
    if (
      !container ||
      !selectedRoomKey ||
      jumpTargetMessageId ||
      hasNewerMessages ||
      isSearchMode ||
      isLoadingMessages ||
      restoredRoomRef.current === selectedRoomKey
    ) {
      return;
    }

    container.scrollTop = container.scrollHeight;
    restoredRoomRef.current = selectedRoomKey;
  }, [
    selectedRoomKey,
    isSearchMode,
    isLoadingMessages,
    jumpTargetMessageId,
    hasNewerMessages,
  ]);

  useEffect(() => {
    if (!jumpTargetMessageId) return;

    requestAnimationFrame(() => {
      const target = document.getElementById(`msg-${jumpTargetMessageId}`);
      target?.scrollIntoView({ block: "start" });
      clearJumpTarget();
    });
  }, [jumpTargetMessageId, clearJumpTarget]);

  useEffect(() => {
    const container = scrollContainerRef.current;
    if (!container || scrollToBottomToken === 0) return;

    requestAnimationFrame(() => {
      container.scrollTo({ top: container.scrollHeight, behavior: "smooth" });
    });
  }, [scrollToBottomToken]);

  const handleScroll = (event: React.UIEvent<HTMLDivElement>) => {
    const target = event.currentTarget;

    if (
      target.scrollTop > TOP_LOAD_THRESHOLD ||
      isLoadingMore ||
      isFetchingOlderRef.current ||
      !hasMoreHistory ||
      !selectedRoom
    ) {
      if (
        hasNewerMessages &&
        !isLoadingNewer &&
        target.scrollHeight - target.scrollTop - target.clientHeight <
          TOP_LOAD_THRESHOLD
      ) {
        void loadNewerMessages(selectedRoom);
      }
      return;
    }

    void triggerLoadMore(target, selectedRoom);
  };

  const triggerLoadMore = async (
    container: HTMLDivElement,
    room: ChatRoom,
  ) => {
    isFetchingOlderRef.current = true;
    const previousScrollHeight = container.scrollHeight;
    const previousScrollTop = container.scrollTop;

    await loadOlderMessages(room);

    requestAnimationFrame(() => {
      const nextContainer = scrollContainerRef.current;
      if (nextContainer) {
        const heightDelta = nextContainer.scrollHeight - previousScrollHeight;
        nextContainer.scrollTop = previousScrollTop + heightDelta;
      }
      isFetchingOlderRef.current = false;
    });
  };

  if (!selectedRoom) {
    return (
      <div className="flex flex-1 items-center justify-center px-6 text-center text-sm text-muted-foreground">
        暂无聊天对象。请先启动后端并同步 chat room 数据。
      </div>
    );
  }

  if (isSearchMode) {
    return (
      <div className="flex-1 overflow-y-auto p-4 space-y-4 scrollbar-thin">
        <div className="text-xs text-muted-foreground">
          搜索结果来自后端。点击任意结果跳转到对应时间。
        </div>
        {searchResults.length === 0 ? (
          <div className="py-10 text-center text-sm text-muted-foreground">
            没有匹配结果
          </div>
        ) : (
          searchResults.map((message) => (
            <button
              type="button"
              key={message.id}
              onClick={() => void jumpToMessage(message, selectedRoom)}
              className="block w-full rounded-md text-left transition-colors hover:bg-accent/50"
            >
              <MessageBubble message={message} room={selectedRoom} userProfile={userProfile} />
            </button>
          ))
        )}
      </div>
    );
  }

  return (
    <div
      ref={scrollContainerRef}
      onScroll={handleScroll}
      className="flex-1 overflow-y-auto overflow-x-hidden scrollbar-thin px-2 relative"
      style={{ scrollBehavior: "auto" }}
    >
      <div
        className={`transition-all duration-300 overflow-hidden ${isLoadingMore ? "h-10 opacity-100" : "h-0 opacity-0"}`}
      >
        <div className="w-full h-full flex justify-center items-center">
          <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          <span className="text-xs text-muted-foreground ml-2">
            加载历史消息...
          </span>
        </div>
      </div>

      {isLoadingMessages ? (
        <div className="flex h-full items-center justify-center">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <div className="flex flex-col justify-end min-h-full">
          {!hasMoreHistory && messageGroups.length > 0 && (
            <div className="text-xs text-center text-muted-foreground py-4 opacity-50">
              没有更多历史记录了
            </div>
          )}

          {messageGroups.length === 0 && (
            <div className="py-10 text-center text-sm text-muted-foreground">
              当前聊天对象暂无消息
            </div>
          )}

          {messageGroups.map((group) => (
            <div key={group.date}>
              <div className="sticky top-0 z-10 flex justify-center mb-4 py-2 pointer-events-none">
                <span className="pointer-events-auto">
                  <DateBadge date={group.date} />
                </span>
              </div>
              <div className="flex flex-col gap-1">
                {group.messages.map((message) => (
                  <div
                    key={message.id}
                    id={`msg-${message.id}`}
                    className="px-2 py-1"
                  >
                    <MessageBubble message={message} room={selectedRoom} userProfile={userProfile} />
                  </div>
                ))}
              </div>
            </div>
          ))}

          {hasNewerMessages && (
            <div className="flex justify-center py-3">
              <button
                type="button"
                onClick={() => void loadNewerMessages(selectedRoom)}
                className="text-xs text-muted-foreground transition-colors hover:text-foreground"
              >
                {isLoadingNewer ? "加载后续消息..." : "加载后续消息"}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

function roomKey(room: ChatRoom) {
  return room.chatRoomId || room.displayName;
}

export default ChatList;
