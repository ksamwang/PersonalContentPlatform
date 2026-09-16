import { ImageResponse } from "next/og";
import { BrandMark } from "../components/BrandMark";

export const size = { width: 180, height: 180 };
export const contentType = "image/png";

export default function AppleIcon() {
  return new ImageResponse(
    (
      <div
        style={{
          alignItems: "center",
          background: "#f4f0e8",
          display: "flex",
          height: "100%",
          justifyContent: "center",
          width: "100%",
        }}
      >
        <BrandMark size={144} />
      </div>
    ),
    size,
  );
}
