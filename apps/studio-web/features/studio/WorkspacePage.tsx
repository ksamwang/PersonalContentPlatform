import { ReactNode } from "react";

type WorkspacePageProps = {
  eyebrow: string;
  title: string;
  description: string;
  actions?: ReactNode;
  className?: string;
  children: ReactNode;
};

export function WorkspacePage({
  eyebrow,
  title,
  description,
  actions,
  className = "",
  children,
}: WorkspacePageProps) {
  return (
    <section className={`workspace-page ${className}`.trim()}>
      <header className="workspace-page__header">
        <div className="workspace-page__intro">
          <span className="eyebrow">{eyebrow}</span>
          <h1>{title}</h1>
          <p>{description}</p>
        </div>
        {actions && <div className="workspace-page__actions">{actions}</div>}
      </header>
      <div className="workspace-page__body">{children}</div>
    </section>
  );
}
