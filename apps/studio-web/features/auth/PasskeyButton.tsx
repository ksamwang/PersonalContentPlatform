"use client";
import { useState } from "react";
import { KeyRound } from "lucide-react";
import { Principal } from "../../lib/api";
import { loginWithPasskey, registerPasskey } from "../../lib/passkey";
export function PasskeyLogin({
  email,
  onAuthenticated,
}: {
  email: string;
  onAuthenticated: (p: Principal) => void;
}) {
  const [busy, setBusy] = useState(false);
  return (
    <button
      type="button"
      className="secondary"
      disabled={!email || busy}
      onClick={async () => {
        setBusy(true);
        try {
          onAuthenticated(await loginWithPasskey(email));
        } finally {
          setBusy(false);
        }
      }}
    >
      <KeyRound aria-hidden size={18} />
      {busy ? "正在验证…" : "使用 Passkey"}
    </button>
  );
}
export function AddPasskey() {
  const [status, setStatus] = useState("");
  return (
    <button
      className="secondary"
      onClick={async () => {
        setStatus("正在创建…");
        try {
          await registerPasskey();
          setStatus("Passkey 已添加");
        } catch {
          setStatus("添加失败");
        }
      }}
    >
      <KeyRound aria-hidden size={17} />
      {status || "添加 Passkey"}
    </button>
  );
}
