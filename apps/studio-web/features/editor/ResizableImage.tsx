"use client";

import Image from "@tiptap/extension-image";
import { NodeViewWrapper, ReactNodeViewRenderer, type ReactNodeViewProps } from "@tiptap/react";
import { useEffect, useRef, useState, type KeyboardEvent, type PointerEvent } from "react";

const MIN_IMAGE_WIDTH = 120;

function ResizableImageNodeView({ node, selected, updateAttributes, editor, getPos }: ReactNodeViewProps) {
  const wrapperRef = useRef<HTMLElement | null>(null);
  const [previewWidth, setPreviewWidth] = useState<number | null>(node.attrs.width ?? null);

  useEffect(() => setPreviewWidth(node.attrs.width ?? null), [node.attrs.width]);

  function maximumWidth() {
    return wrapperRef.current?.parentElement?.clientWidth ?? Number.MAX_SAFE_INTEGER;
  }

  function clampWidth(width: number) {
    return Math.min(Math.max(MIN_IMAGE_WIDTH, Math.round(width)), maximumWidth());
  }

  function beginResize(event: PointerEvent<HTMLButtonElement>) {
    event.preventDefault();
    event.stopPropagation();
    const handle = event.currentTarget;
    const startX = event.clientX;
    const startWidth = wrapperRef.current?.querySelector("img")?.getBoundingClientRect().width ?? MIN_IMAGE_WIDTH;
    handle.setPointerCapture(event.pointerId);

    const move = (moveEvent: globalThis.PointerEvent) => {
      setPreviewWidth(clampWidth(startWidth + moveEvent.clientX - startX));
    };
    const finish = (upEvent: globalThis.PointerEvent) => {
      handle.releasePointerCapture(upEvent.pointerId);
      handle.removeEventListener("pointermove", move);
      handle.removeEventListener("pointerup", finish);
      handle.removeEventListener("pointercancel", finish);
      const width = clampWidth(startWidth + upEvent.clientX - startX);
      setPreviewWidth(width);
      updateAttributes({ width });
    };
    handle.addEventListener("pointermove", move);
    handle.addEventListener("pointerup", finish);
    handle.addEventListener("pointercancel", finish);
  }

  function resizeWithKeyboard(event: KeyboardEvent<HTMLButtonElement>) {
    if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") return;
    event.preventDefault();
    const currentWidth = wrapperRef.current?.querySelector("img")?.getBoundingClientRect().width ?? MIN_IMAGE_WIDTH;
    const width = clampWidth(currentWidth + (event.key === "ArrowRight" ? 16 : -16));
    setPreviewWidth(width);
    updateAttributes({ width });
  }

  return (
    <NodeViewWrapper
      ref={wrapperRef}
      className={`resizable-image${selected ? " is-selected" : ""}`}
      style={{ width: previewWidth ? `${previewWidth}px` : "fit-content" }}
      onClick={() => {
        const position = getPos();
        if (typeof position === "number") editor.commands.setNodeSelection(position);
      }}
    >
      <img src={node.attrs.src} alt={node.attrs.alt ?? ""} title={node.attrs.title ?? undefined} draggable={false} />
      {selected && (
        <button
          type="button"
          className="image-resize-handle"
          aria-label="调整图片大小"
          title="拖动或使用左右方向键调整图片大小"
          onPointerDown={beginResize}
          onKeyDown={resizeWithKeyboard}
        />
      )}
    </NodeViewWrapper>
  );
}

export const ResizableImage = Image.extend({
  addAttributes() {
    return {
      ...this.parent?.(),
      width: {
        default: null,
        parseHTML: (element) => {
          const value = element.getAttribute("width") ?? element.style.width;
          const width = Number.parseInt(value, 10);
          return Number.isFinite(width) ? width : null;
        },
        renderHTML: (attributes) =>
          attributes.width
            ? { width: attributes.width, style: `width: ${attributes.width}px; height: auto;` }
            : {},
      },
    };
  },
  addNodeView() {
    return ReactNodeViewRenderer(ResizableImageNodeView);
  },
}).configure({ allowBase64: false });
