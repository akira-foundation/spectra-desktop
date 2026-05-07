import { cn } from '@/lib/utils'

interface ProjectAvatarProps {
  name: string
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

const sizeMap = {
  sm: 'w-5 h-5 text-[10.5px]',
  md: 'w-7 h-7 text-[12px]',
  lg: 'w-10 h-10 text-[15px]',
}

export function ProjectAvatar({ name, size = 'sm', className }: ProjectAvatarProps) {
  const initial = (name?.charAt(0) ?? '?').toUpperCase()
  return (
    <span
      className={cn(
        'inline-flex items-center justify-center rounded bg-primary/15 text-primary font-semibold shrink-0',
        sizeMap[size],
        className,
      )}
    >
      {initial}
    </span>
  )
}
