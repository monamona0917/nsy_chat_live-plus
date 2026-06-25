import {
  Dialog,
  DialogHeader,
  DialogPanel,
  DialogPopup,
  DialogTitle,
} from "@/components/ui/dialog";

interface AboutDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

const Link = ({
  href,
  children,
}: {
  href: string;
  children: React.ReactNode;
}) => (
  <a
    href={href}
    target="_blank"
    rel="noopener noreferrer"
    className="inline-flex items-center gap-1 text-primary underline-offset-4 hover:underline mx-1"
  >
    {children}
  </a>
);

const AboutDialog = ({ isOpen, onClose }: AboutDialogProps) => {
  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>关于</DialogTitle>
        </DialogHeader>
        <DialogPanel>
          <p>
            本页面基于
            <Link href="https://github.com/Chilfish/replive-oyu">
              https://github.com/Chilfish/replive-oyu
            </Link>
            二次开发。感谢。
          </p>
        </DialogPanel>
      </DialogPopup>
    </Dialog>
  );
};

export default AboutDialog;
