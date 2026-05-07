import { Lightbulb } from 'lucide-react';
import { ReactNode } from 'react';

interface TipCardProps {
  title: string;
  description: string | ReactNode;
  variant?: 'info' | 'tip' | 'warning';
}

export function TipCard({
  title,
  description,
  variant = 'tip'
}: TipCardProps) {
  const variantStyles = {
    info: {
      bg: 'bg-blue-50 dark:bg-blue-950/30',
      border: 'border-blue-200 dark:border-blue-800',
      icon: 'text-blue-600 dark:text-blue-400',
      title: 'text-blue-900 dark:text-blue-200',
      text: 'text-blue-800 dark:text-blue-300'
    },
    tip: {
      bg: 'bg-amber-50 dark:bg-amber-950/30',
      border: 'border-amber-200 dark:border-amber-800',
      icon: 'text-amber-600 dark:text-amber-400',
      title: 'text-amber-900 dark:text-amber-200',
      text: 'text-amber-800 dark:text-amber-300'
    },
    warning: {
      bg: 'bg-red-50 dark:bg-red-950/30',
      border: 'border-red-200 dark:border-red-800',
      icon: 'text-red-600 dark:text-red-400',
      title: 'text-red-900 dark:text-red-200',
      text: 'text-red-800 dark:text-red-300'
    }
  };

  const style = variantStyles[variant];

  return (
    <div
      className={`rounded-xl border-2 ${style.border} ${style.bg} px-4 py-3`}
    >
      <div className="flex gap-3">
        <div className={`mt-0.5 flex-shrink-0 ${style.icon}`}>
          <Lightbulb className="h-5 w-5" />
        </div>
        <div>
          <p className={`font-semibold text-sm ${style.title}`}>
            {title}
          </p>
          <p className={`text-sm mt-1 ${style.text}`}>
            {description}
          </p>
        </div>
      </div>
    </div>
  );
}
