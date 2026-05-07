export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

interface HttpMethodColor {
  badge: string
  text: string
}

const HTTP_METHOD_COLORS: Record<HttpMethod, HttpMethodColor> = {
  'GET': {
    badge: 'bg-emerald-500/20',
    text: 'text-emerald-400'
  },
  'POST': {
    badge: 'bg-blue-500/20',
    text: 'text-blue-400'
  },
  'PUT': {
    badge: 'bg-amber-500/20',
    text: 'text-amber-400'
  },
  'PATCH': {
    badge: 'bg-orange-500/20',
    text: 'text-orange-400'
  },
  'DELETE': {
    badge: 'bg-red-500/20',
    text: 'text-red-400'
  }
}

export function useHttpMethod() {
  const getMethodColor = (method: string): string => {
    const upperMethod = method.toUpperCase() as HttpMethod
    const colors = HTTP_METHOD_COLORS[upperMethod]
    
    if (!colors) {
      return 'bg-gray-500/20 text-gray-400'
    }
    
    return `${colors.badge} ${colors.text}`
  }

  const getMethodBadgeColor = (method: string): string => {
    const upperMethod = method.toUpperCase() as HttpMethod
    return HTTP_METHOD_COLORS[upperMethod]?.badge || 'bg-gray-500/20'
  }

  const getMethodTextColor = (method: string): string => {
    const upperMethod = method.toUpperCase() as HttpMethod
    return HTTP_METHOD_COLORS[upperMethod]?.text || 'text-gray-400'
  }

  return {
    getMethodColor,
    getMethodBadgeColor,
    getMethodTextColor,
    HTTP_METHOD_COLORS
  }
}
