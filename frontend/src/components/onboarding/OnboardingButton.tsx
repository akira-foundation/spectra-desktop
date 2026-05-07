import { Button } from '@/components/ui/button';
import { ButtonHTMLAttributes, ReactNode } from 'react';

interface OnboardingButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary';
  children: ReactNode;
}

export function OnboardingButton({ 
  variant = 'primary', 
  children, 
  className = '',
  ...props 
}: OnboardingButtonProps) {
  return (
    <Button
      variant={variant === 'primary' ? 'gradient' : 'outline'}
      size="lg"
      className={`min-w-[200px] ${className}`}
      {...props}
    >
      {children}
    </Button>
  );
}
