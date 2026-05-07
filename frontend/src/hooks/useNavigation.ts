import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface ResponseHistoryItem {
  id: string
  timestamp: number
  response: any
}

interface Metrics {
  totalRequests: number
  successfulRequests: number
  clientErrors: number
  serverErrors: number
  responseTimes: number[]
  lastRequestTime: number | null
  requestHistory: any[]
}

interface NavigationState {
  selectedEndpoint: string | null
  setSelectedEndpoint: (endpoint: string | null) => void
  response: any | null
  setResponse: (response: any | null) => void
  expandedGroups: Set<string>
  toggleGroup: (groupName: string) => void
  bodyMode: 'json' | 'form'
  setBodyMode: (mode: 'json' | 'form') => void
  responseHistory: Record<string, ResponseHistoryItem[]>
  addToHistory: (endpointKey: string, response: any) => void
  clearHistory: (endpointKey?: string) => void
  metrics: Metrics
  recordRequest: (status: number, timeMs: number, method: string, url: string) => void
  resetMetrics: () => void
}

export const useNavigation = create<NavigationState>()(
  persist(
    (set) => ({
      selectedEndpoint: null,
      setSelectedEndpoint: (endpoint) => set({ selectedEndpoint: endpoint }),
      response: null,
      setResponse: (response) => set({ response }),
      expandedGroups: new Set<string>(['auth']),
      toggleGroup: (groupName) =>
        set((state) => {
          const newExpanded = new Set(state.expandedGroups)
          if (newExpanded.has(groupName)) {
            newExpanded.delete(groupName)
          } else {
            newExpanded.add(groupName)
          }
          return { expandedGroups: newExpanded }
        }),
      bodyMode: 'json',
      setBodyMode: (mode) => set({ bodyMode: mode }),
      responseHistory: {},
      addToHistory: (endpointKey, response) =>
        set((state) => {
          const newItem: ResponseHistoryItem = {
            id: `${endpointKey}-${Date.now()}`,
            timestamp: Date.now(),
            response,
          }
          
          const currentHistory = state.responseHistory[endpointKey] || []
          // Keep last 10 items per endpoint
          const updatedHistory = [newItem, ...currentHistory].slice(0, 10)
          
          return {
            responseHistory: {
              ...state.responseHistory,
              [endpointKey]: updatedHistory,
            },
          }
        }),
      clearHistory: (endpointKey) =>
        set((state) => {
          if (endpointKey) {
            const { [endpointKey]: _, ...rest } = state.responseHistory
            return { responseHistory: rest }
          }
          return { responseHistory: {} }
        }),
      metrics: {
        totalRequests: 0,
        successfulRequests: 0,
        clientErrors: 0,
        serverErrors: 0,
        responseTimes: [],
        requestHistory: [],
        lastRequestTime: null,
      },
      recordRequest: (status, timeMs, method, url) =>
        set((state) => {
          const isSuccess = status >= 200 && status < 300
          const isClientError = status >= 400 && status < 500
          const isServerError = status >= 500
          const newResponseTimes = [...state.metrics.responseTimes, timeMs].slice(-20)

          const newRequestHistoryItem = {
            method,
            url,
            status,
            responseTime: timeMs,
            timestamp: Date.now(),
          }

          const newRequestHistory = [...state.metrics.requestHistory, newRequestHistoryItem].slice(-50)

          return {
            metrics: {
              totalRequests: state.metrics.totalRequests + 1,
              successfulRequests: state.metrics.successfulRequests + (isSuccess ? 1 : 0),
              clientErrors: state.metrics.clientErrors + (isClientError ? 1 : 0),
              serverErrors: state.metrics.serverErrors + (isServerError ? 1 : 0),
              responseTimes: newResponseTimes,
              requestHistory: newRequestHistory,
              lastRequestTime: Date.now(),
            },
          }
        }),
      resetMetrics: () =>
        set({
          metrics: {
            totalRequests: 0,
            successfulRequests: 0,
            clientErrors: 0,
            serverErrors: 0,
            responseTimes: [],
            requestHistory: [],
            lastRequestTime: null,
          },
        }),
    }),
    {
      name: 'spectra-navigation',
      version: 1,
      storage: {
        getItem: (name) => {
          try {
            const item = localStorage.getItem(name)
            if (!item) return null
            const parsed = JSON.parse(item)
            
            // Check version compatibility
            if (parsed.version !== 1) {
              localStorage.removeItem(name)
              return null
            }
            
            return {
              state: {
                ...parsed.state,
                expandedGroups: new Set(parsed.state.expandedGroups || ['auth']),
                responseHistory: parsed.state.responseHistory || {},
                metrics: {
                  ...parsed.state.metrics,
                  requestHistory: Array.isArray(parsed.state.metrics?.requestHistory) ? parsed.state.metrics.requestHistory : [],
                },
              },
              version: parsed.version,
            }
          } catch (error) {
            console.error('Error loading navigation state:', error)
            localStorage.removeItem(name)
            return null
          }
        },
        setItem: (name, value) => {
          localStorage.setItem(
            name,
            JSON.stringify({
              state: {
                ...value.state,
                expandedGroups: Array.from(value.state.expandedGroups),
              },
              version: value.version,
            })
          )
        },
        removeItem: (name) => localStorage.removeItem(name),
      },
      partialize: (state) => ({
        selectedEndpoint: state.selectedEndpoint,
        expandedGroups: state.expandedGroups,
        bodyMode: state.bodyMode,
        responseHistory: state.responseHistory,
        metrics: state.metrics,
      }),
    }
  )
)
