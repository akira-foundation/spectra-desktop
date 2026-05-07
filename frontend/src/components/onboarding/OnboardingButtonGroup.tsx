import { ReactNode } from 'react';

interface OnboardingButtonGroupProps {
  children: ReactNode;
}

export function OnboardingButtonGroup({ children }: OnboardingButtonGroupProps) {
  return (
    <div className="flex gap-3 justify-center mt-4">
      {children}
    </div>
  );
}
