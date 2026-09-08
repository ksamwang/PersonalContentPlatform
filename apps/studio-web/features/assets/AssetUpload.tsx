"use client";
import { ChangeEvent, useState } from "react";
import { ImagePlus } from "lucide-react";
import { api, Asset } from "../../lib/api";
export function AssetUpload({
  workspace,
  onUploaded,
}: {
  workspace: string;
  onUploaded?: (a: Asset) => void;
}) {
  const [status, setStatus] = useState("");
  async function change(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    setStatus("上传中…");
    try {
      const asset = await api.uploadAsset(workspace, file);
      setStatus("上传完成");
      onUploaded?.(asset);
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "上传失败");
    } finally {
      event.target.value = "";
    }
  }
  return (
    <label className="secondary asset-upload">
      <ImagePlus aria-hidden size={17} />
      <span>{status || "上传图片"}</span>
      <input
        className="sr-only"
        type="file"
        accept="image/*"
        onChange={change}
      />
    </label>
  );
}
