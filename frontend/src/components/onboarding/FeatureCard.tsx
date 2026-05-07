import { ReactNode } from 'react';

interface FeatureCardProps {
  icon: ReactNode;
  title: string;
  description: string;
  variant?: 'default' | 'interactive';
  onClick?: () => void;
  selected?: boolean;
}

export function FeatureCard({
  icon,
  title,
  description,
  variant = 'default',
  onClick,
  selected = false
}: FeatureCardProps) {
  return (
    <div
      onClick={onClick}
      className={`
        relative rounded-2xl border-2 p-6 transition-all duration-300
        ${
          variant === 'interactive'
            ? 'cursor-pointer group hover:border-violet-400 dark:hover:border-violet-500'
            : ''
        }
        ${
          selected
            ? 'border-violet-500 shadow-lg shadow-violet-200 dark:shadow-violet-900/20'
            : 'border-slate-200 dark:border-white/10 hover:border-slate-300 dark:hover:border-white/20'
        }
      `}
    >
      <div className="mb-4 inline-flex rounded-xl bg-gradient-to-br from-violet-100 to-purple-100 dark:from-violet-900/40 dark:to-purple-900/40 p-3">
        <div className="text-2xl text-violet-600 dark:text-violet-400">
          {icon}
        </div>
      </div>

      <h3 className="mb-2 text-lg font-semibold text-slate-900 dark:text-white">
        {title}
      </h3>

      <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
        {description}
      </p>

      {variant === 'interactive' && selected && (
        <div className="mt-4 flex items-center gap-2 text-sm font-medium text-violet-600 dark:text-violet-400">
          <div className="flex h-5 w-5 items-center justify-center rounded-full bg-violet-600 dark:bg-violet-500">
            <svg
              className="h-3 w-3 text-white"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fillRule="evenodd"
                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                clipRule="evenodd"
              />
            </svg>
          </div>
          Selected
        </div>
      )}
    </div>
  );
}
