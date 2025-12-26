import type { ButtonHTMLAttributes } from 'react';

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary';
};

export function Button({ variant = 'primary', ...props }: ButtonProps) {
  const className =
    variant === 'primary'
      ? 'ui-button ui-button--primary'
      : 'ui-button ui-button--secondary';

  return <button className={className} {...props} />;
}
