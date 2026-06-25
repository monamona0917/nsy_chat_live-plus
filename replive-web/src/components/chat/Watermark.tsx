import type React from "react";

interface WatermarkProps {
  /** The text to display in the watermark. */
  text: string;
}

/**
 * A component to display a tiled, rotated watermark over the content.
 * Uses mix-blend-mode: difference with a fixed mid-gray SVG fill so the
 * watermark is visible on both light and dark backgrounds without relying
 * on CSS variable inheritance into SVG data URIs.
 */
const Watermark: React.FC<WatermarkProps> = ({ text }) => {
  const svgWidth = 400;
  const svgHeight = 400;

  // SVG 使用固定颜色 #808080（中灰色），不依赖 CSS 变量透传
  const svgString = `
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="${svgWidth}"
      height="${svgHeight}"
      viewBox="0 0 ${svgWidth} ${svgHeight}"
    >
      <text
        x="50%"
        y="50%"
        text-anchor="middle"
        dominant-baseline="middle"
        transform="rotate(-30, ${svgWidth / 2}, ${svgHeight / 2})"
        fill="#808080"
        style="
          font-size: 16px;
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans', sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol', 'Noto Color Emoji';
          font-weight: 600;
        "
      >
        ${text}
      </text>
    </svg>
  `;

  const dataUrl = `url("data:image/svg+xml,${encodeURIComponent(
    svgString.replace(/[\n\s]+/g, " "),
  )}")`;

  return (
    <div
      className="pointer-events-none fixed inset-0 z-30"
      style={{
        backgroundImage: dataUrl,
        backgroundRepeat: "repeat",
        mixBlendMode: "difference",
        opacity: 0.15,
      }}
      aria-hidden="true"
    />
  );
};

export default Watermark;
