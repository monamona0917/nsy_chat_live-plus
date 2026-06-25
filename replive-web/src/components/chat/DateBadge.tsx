import { format } from "date-fns";

import { zhCN } from "date-fns/locale";
import { Badge } from "../ui/badge";

interface DateBadgeProps {
  date: string;
}

const DateBadge = ({ date }: DateBadgeProps) => {
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return format(date, "yyyy年M月d日 EEEE", { locale: zhCN });
  };

  return (
    <div className="flex justify-center my-4">
      <Badge
        variant="secondary"
        className="px-3 py-1 text-xs text-muted-foreground bg-background/80 backdrop-blur-sm"
      >
        {formatDate(date)}
      </Badge>
    </div>
  );
};

export default DateBadge;
