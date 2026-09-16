type Props = {
  className?: string;
  size?: number;
  title?: string;
};

export function BrandMark({ className, size = 32, title }: Props) {
  return (
    <svg
      className={className}
      width={size}
      height={size}
      viewBox="0 0 40 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      role={title ? "img" : undefined}
      aria-hidden={title ? undefined : true}
      aria-label={title}
    >
      <circle cx="20" cy="20" r="15" fill="#171416" />
      <ellipse
        cx="20"
        cy="20"
        rx="17.2"
        ry="7.2"
        transform="rotate(-18 20 20)"
        stroke="#F4F0E8"
        strokeWidth="1.6"
      />
      <path
        d="M7.9 27.5C14.8 31.2 25.7 29.6 31.9 23.4"
        stroke="#D82F76"
        strokeWidth="2"
        strokeLinecap="round"
      />
      <circle cx="29.6" cy="13.4" r="2.7" fill="#D82F76" />
      <circle cx="20" cy="20" r="4.2" fill="#F4F0E8" />
      <circle cx="20" cy="20" r="2.2" fill="#171416" />
    </svg>
  );
}
