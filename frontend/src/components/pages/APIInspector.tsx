import { useMemo, useState } from 'react'
import { useUIStore } from '@/store/uiStore'
import {
  EndpointList,
  AuthenticationDrawer,
  RequestPanel,
  ResponsePanel,
  EndpointHeader,
} from '@/components/api-inspector'
import { EndpointMetadata } from '@/components/api-inspector/EndpointMetadata'
import { EndpointEmptyState } from '@/components/api-inspector/EndpointEmptyState'
import { ResponseEmptyState } from '@/components/api-inspector/ResponseEmptyState'

interface MockEndpoint {
  method: string
  name: string
  path: string
  tag: string
  controller?: string
  sourceFile?: string
  sourceLine?: number
  middleware?: string[]
  authRequired?: boolean
}

interface MockCategory {
  category: string
  count: number
  items: MockEndpoint[]
}

const endpoints: MockCategory[] = [
  {
    category: 'Auth',
    count: 6,
    items: [
      { method: 'POST', name: 'Forgot Password', path: 'api/auth/forgot-password', tag: 'password.email' },
      { method: 'POST', name: 'Login', path: 'api/auth/login', tag: 'auth.login' },
      { method: 'POST', name: 'Logout', path: 'api/auth/logout', tag: 'auth.logout' },
      { method: 'PUT', name: 'Update Password', path: 'api/auth/password', tag: 'password.update' },
      {
        method: 'POST',
        name: 'Register',
        path: 'api/auth/register',
        tag: 'auth.register',
        controller: 'App\\Http\\Controllers\\Auth\\RegisterController@store',
        sourceFile: 'routes/api.php',
        sourceLine: 42,
        middleware: ['api', 'throttle:60,1', 'guest'],
        authRequired: false,
      },
      { method: 'POST', name: 'Reset Password', path: 'api/auth/reset-password', tag: 'password.store' },
    ],
  },
  {
    category: 'Users',
    count: 5,
    items: [
      { method: 'GET', name: 'List Users', path: 'api/users', tag: 'users.index' },
      { method: 'POST', name: 'Create User', path: 'api/users', tag: 'users.store' },
      { method: 'GET', name: 'Show User', path: 'api/users/{id}', tag: 'users.show' },
      { method: 'PUT', name: 'Update User', path: 'api/users/{id}', tag: 'users.update' },
      { method: 'DELETE', name: 'Delete User', path: 'api/users/{id}', tag: 'users.destroy' },
    ],
  },
  {
    category: 'Posts',
    count: 6,
    items: [
      { method: 'GET', name: 'List Posts', path: 'api/posts', tag: 'posts.index' },
      { method: 'POST', name: 'Create Post', path: 'api/posts', tag: 'posts.store' },
      { method: 'GET', name: 'Show Post', path: 'api/posts/{id}', tag: 'posts.show' },
      { method: 'PUT', name: 'Update Post', path: 'api/posts/{id}', tag: 'posts.update' },
      { method: 'DELETE', name: 'Delete Post', path: 'api/posts/{id}', tag: 'posts.destroy' },
      { method: 'POST', name: 'Publish Post', path: 'api/posts/{id}/publish', tag: 'posts.publish' },
    ],
  },
  {
    category: 'Comments',
    count: 4,
    items: [
      { method: 'GET', name: 'List Comments', path: 'api/comments', tag: 'comments.index' },
      { method: 'POST', name: 'Create Comment', path: 'api/comments', tag: 'comments.store' },
      { method: 'PUT', name: 'Update Comment', path: 'api/comments/{id}', tag: 'comments.update' },
      { method: 'DELETE', name: 'Delete Comment', path: 'api/comments/{id}', tag: 'comments.destroy' },
    ],
  },
  {
    category: 'Media',
    count: 3,
    items: [
      { method: 'POST', name: 'Upload File', path: 'api/media/upload', tag: 'media.upload' },
      { method: 'GET', name: 'List Media', path: 'api/media', tag: 'media.index' },
      { method: 'DELETE', name: 'Delete Media', path: 'api/media/{id}', tag: 'media.destroy' },
    ],
  },
]

export function APIInspector() {
  const [selectedTag, setSelectedTag] = useState<string>('auth.register')
  const [hasResponse, setHasResponse] = useState(true)
  const activeAuthMethod = useUIStore((state) => state.activeAuthMethod)
  const setActiveAuthMethod = useUIStore((state) => state.setActiveAuthMethod)

  const requestBody = {
    name: 'Bob Wilson',
    email: 'user2059@example.com',
    password: 'password205',
    password_confirmation: 'password205',
  }

  const responseData = {
    id: 63,
    name: 'John Doe',
    email: 'user1379@example.com',
    email_verified_at: null,
    created_at: '2025-11-30T18:27:03+00:00',
    updated_at: '2025-11-30T18:27:03+00:00',
  }

  const decoratedEndpoints = useMemo(
    () =>
      endpoints.map((cat) => ({
        ...cat,
        items: cat.items.map((item) => ({ ...item, active: item.tag === selectedTag })),
      })),
    [selectedTag],
  )

  const selected = useMemo(() => {
    for (const cat of endpoints) {
      const found = cat.items.find((i) => i.tag === selectedTag)
      if (found) return found
    }
    return null
  }, [selectedTag])

  const handleExecute = () => {
    setHasResponse(true)
    console.log('Execute request')
  }

  return (
    <div className="h-full flex overflow-hidden">
      <EndpointList endpoints={decoratedEndpoints} onSelectEndpoint={setSelectedTag} />

      <div className="flex-1 flex flex-col overflow-hidden">
        <AuthenticationDrawer
          activeMethod={activeAuthMethod}
          onMethodChange={setActiveAuthMethod}
        />

        {!selected ? (
          <EndpointEmptyState />
        ) : (
          <>
            <EndpointHeader
              method={selected.method}
              path={selected.path}
              statusCode={hasResponse ? 201 : 0}
              responseTime={hasResponse ? '260ms' : '—'}
              responseSize={hasResponse ? '0.16KB' : '—'}
            />

            <EndpointMetadata
              controller={selected.controller}
              sourceFile={selected.sourceFile}
              sourceLine={selected.sourceLine}
              middleware={selected.middleware}
              authRequired={selected.authRequired}
            />

            <div className="flex-1 grid grid-cols-2 overflow-hidden">
              <RequestPanel requestBody={requestBody} onExecute={handleExecute} />
              {hasResponse ? (
                <ResponsePanel responseData={responseData} />
              ) : (
                <div className="bg-background flex items-center justify-center">
                  <ResponseEmptyState />
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
