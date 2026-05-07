interface OnboardingProgressProps {
  currentStep: number;
  totalSteps: number;
  stepLabels?: string[];
}

export function OnboardingProgress({
  currentStep,
  totalSteps
}: OnboardingProgressProps) {
  return (
    <div className="flex justify-center gap-2">
      {Array.from({ length: totalSteps }).map((_, index) => (
        <div
          key={index}
          className={`h-1.5 rounded-full transition-all ${
            index === currentStep
              ? 'w-8 bg-primary'
              : index < currentStep
                ? 'w-1.5 bg-primary/50'
                : 'w-1.5 bg-muted'
          }`}
        />
      ))}
    </div>
  );
}
