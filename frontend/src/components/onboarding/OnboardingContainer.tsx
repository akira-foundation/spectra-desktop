import { ReactNode } from 'react';

interface OnboardingContainerProps {
  children: ReactNode;
  variant?: 'default' | 'centered' | 'full';
}

export function OnboardingContainer({
  children,
  variant = 'centered'
}: OnboardingContainerProps) {
  const variantClasses = {
    default: 'min-h-screen w-full flex items-center justify-center p-8',
    centered: 'min-h-screen w-full flex items-center justify-center p-8',
    full: 'w-full h-screen flex items-center justify-center'
  };

  return (
    <div className={variantClasses[variant]}>
      <div className="w-full max-w-3xl">
        <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
          {children}
        </div>
      </div>
    </div>
  );
}
