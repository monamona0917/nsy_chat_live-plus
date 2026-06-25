import { Search, X } from "lucide-react";
import { useEffect, useState } from "react";
import useChatStore from "../../stores/chat-store";
import { Button } from "../ui/button";
import { Input } from "../ui/input";

interface SearchBarProps {
  isOpen: boolean;
  onClose: () => void;
}

const SearchBar = ({ isOpen, onClose }: SearchBarProps) => {
  const { searchQuery, setSearchQuery, searchMessages, searchResults, isSearching, clearSearch } =
    useChatStore();
  const [localQuery, setLocalQuery] = useState(searchQuery);

  useEffect(() => {
    if (isOpen) {
      setLocalQuery(searchQuery);
    }
  }, [isOpen, searchQuery]);

  useEffect(() => {
    const timer = setTimeout(() => {
      setSearchQuery(localQuery);
      void searchMessages(localQuery);
    }, 300);

    return () => clearTimeout(timer);
  }, [localQuery, setSearchQuery, searchMessages]);

  const handleClose = () => {
    clearSearch();
    setLocalQuery("");
    onClose();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") {
      handleClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed top-0 left-0 right-0 z-50 bg-background/95 backdrop-blur-sm border-b shadow-sm">
      <div className="container mx-auto px-4 py-3">
        <div className="flex items-center space-x-3">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              value={localQuery}
              onChange={(e) => setLocalQuery(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="按关键字搜索聊天记录..."
              className="pl-10 pr-4"
              autoFocus
            />
          </div>

          <div className="flex items-center space-x-2">
            {searchResults.length > 0 && (
              <span className="text-sm text-muted-foreground whitespace-nowrap">
                找到 {searchResults.length} 条消息
              </span>
            )}
            {isSearching && (
              <span className="text-sm text-muted-foreground whitespace-nowrap">
                搜索中...
              </span>
            )}

            <Button variant="ghost" size="sm" onClick={handleClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SearchBar;
