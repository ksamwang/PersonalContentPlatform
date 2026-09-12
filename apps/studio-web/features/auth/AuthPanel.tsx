"use client";
import { FormEvent, useEffect, useState } from "react";
import { api, Principal } from "../../lib/api";
import { PasskeyLogin } from "./PasskeyButton";
export function AuthPanel({
  onAuthenticated,
}: {
  onAuthenticated: (p: Principal) => void;
}) {
  const [setupRequired, setSetupRequired] = useState<boolean>(),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false),
    [email, setEmail] = useState("");
  const checking = setupRequired === undefined;
  const setup = setupRequired === true;
  useEffect(() => {
    api.setupStatus().then(value => setSetupRequired(value.required)).catch(() => setSetupRequired(false));
  }, []);
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    const data = new FormData(e.currentTarget);
    const input = {
      email: data.get("email"),
      password: data.get("password"),
      display_name: data.get("display_name"),
    };
    try {
      onAuthenticated(setup ? await api.setup(input) : await api.login(input));
    } catch (err) {
      setError(err instanceof Error ? err.message : "无法登录");
    } finally {
      setBusy(false);
    }
  }
  return (
    <main id="main" className="auth-shell">
      <section className="auth-story">
        <span className="eyebrow">PERSONAL CONTENT PLATFORM</span>
        <h1>
          让每一次记录，
          <br />
          在未来继续生长。
        </h1>
        <p>捕获灵感、形成内容、连接知识，并以自己的方式发布。</p>
      </section>
      <section className="auth-card">
        <div className="auth-card-heading">
          <span className="eyebrow">{checking ? "OPENING" : setup ? "FIRST SETUP" : "WELCOME BACK"}</span>
          <h2>{checking ? "正在准备登录" : setup ? "创建个人空间" : "登录 Studio"}</h2>
          <p>{checking ? "正在检查当前实例的初始化状态…" : setup ? "完成一次初始化后，这个入口将自动隐藏。" : "继续整理、连接并发布你的内容。"}</p>
        </div>
        {!checking && <form onSubmit={submit}>
          {setup && (
            <label>
              显示名称
              <input name="display_name" autoComplete="name" required />
            </label>
          )}
          <label>
            邮箱
            <input
              name="email"
              type="email"
              autoComplete="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              required
            />
          </label>
          <label>
            密码
            <input
              name="password"
              type="password"
              autoComplete={setup ? "new-password" : "current-password"}
              minLength={10}
              required
            />
            <small>至少 10 个字符</small>
          </label>
          {error && (
            <p className="form-error" role="alert">
              {error}
            </p>
          )}
          <button className="primary" disabled={busy}>
            {busy ? "处理中…" : setup ? "创建个人空间" : "进入 Studio"}
          </button>
        </form>}
        {setupRequired === false && (
          <PasskeyLogin email={email} onAuthenticated={onAuthenticated} />
        )}
        {setupRequired === false && <p className="auth-note">Passkey 可在登录后添加到当前设备。</p>}
      </section>
    </main>
  );
}
