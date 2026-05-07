import { useState } from 'react'
import { useUIStore } from '@/store/uiStore'
import {
  EndpointList,
  AuthenticationDrawer,
  RequestPanel,
  ResponsePanel,
  EndpointHeader,
} from '@/components/api-inspector'

export function APIInspector() {
  const [selectedEndpoint, setSelectedEndpoint] = useState('register')
  const activeAuthMethod = useUIStore((state) => state.activeAuthMethod)
  const setActiveAuthMethod = useUIStore((state) => state.setActiveAuthMethod)
  
  const [requestBody] = useState({
    name: "Bob Wilson",
    email: "user2059@example.com",
    password: "password205",
    password_confirmation: "password205"
  })
  
  const [responseData] = useState({
    id: 63,
    name: "John Doe",
    email: "user1379@example.com",
    email_verified_at: null,
    created_at: "2025-11-30T18:27:03+00:00",
    updated_at: "2025-11-30T18:27:03+00:00"
  })

  const endpoints = [
    {
      category: "Auth",
      count: 6,
      items: [
        { method: "POST", name: "Forgot Password", path: "api/auth/forgot-password", tag: "password.email", active: false },
        { method: "POST", name: "Login", path: "api/auth/login", tag: "auth.login", active: false },
        { method: "POST", name: "Logout", path: "api/auth/logout", tag: "auth.logout", active: false },
        { method: "PUT", name: "Update Password", path: "api/auth/password", tag: "password.update", active: false },
        { method: "POST", name: "Register", path: "api/auth/register", tag: "auth.register", active: true },
        { method: "POST", name: "Reset Password", path: "api/auth/reset-password", tag: "password.store", active: false },
      ]
    },
    {
      category: "Users",
      count: 5,
      items: [
        { method: "GET", name: "List Users", path: "api/users", tag: "users.index", active: false },
        { method: "POST", name: "Create User", path: "api/users", tag: "users.store", active: false },
        { method: "GET", name: "Show User", path: "api/users/{id}", tag: "users.show", active: false },
        { method: "PUT", name: "Update User", path: "api/users/{id}", tag: "users.update", active: false },
        { method: "DELETE", name: "Delete User", path: "api/users/{id}", tag: "users.destroy", active: false },
      ]
    },
    {
      category: "Posts",
      count: 6,
      items: [
        { method: "GET", name: "List Posts", path: "api/posts", tag: "posts.index", active: false },
        { method: "POST", name: "Create Post", path: "api/posts", tag: "posts.store", active: false },
        { method: "GET", name: "Show Post", path: "api/posts/{id}", tag: "posts.show", active: false },
        { method: "PUT", name: "Update Post", path: "api/posts/{id}", tag: "posts.update", active: false },
        { method: "DELETE", name: "Delete Post", path: "api/posts/{id}", tag: "posts.destroy", active: false },
        { method: "POST", name: "Publish Post", path: "api/posts/{id}/publish", tag: "posts.publish", active: false },
      ]
    },
    {
      category: "Comments",
      count: 4,
      items: [
        { method: "GET", name: "List Comments", path: "api/comments", tag: "comments.index", active: false },
        { method: "POST", name: "Create Comment", path: "api/comments", tag: "comments.store", active: false },
        { method: "PUT", name: "Update Comment", path: "api/comments/{id}", tag: "comments.update", active: false },
        { method: "DELETE", name: "Delete Comment", path: "api/comments/{id}", tag: "comments.destroy", active: false },
      ]
    },
    {
      category: "Media",
      count: 3,
      items: [
        { method: "POST", name: "Upload File", path: "api/media/upload", tag: "media.upload", active: false },
        { method: "GET", name: "List Media", path: "api/media", tag: "media.index", active: false },
        { method: "DELETE", name: "Delete Media", path: "api/media/{id}", tag: "media.destroy", active: false },
      ]
    },
  ]

  const handleExecute = () => {
    console.log('Execute request')
  }

  return (
    <div className="h-full flex overflow-hidden">
      {/* Left Sidebar - Endpoints */}
      <EndpointList
        endpoints={endpoints}
        onSelectEndpoint={setSelectedEndpoint}
      />

      {/* Main Content - Right Side */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Authentication Drawer */}
        <AuthenticationDrawer
          activeMethod={activeAuthMethod}
          onMethodChange={setActiveAuthMethod}
        />

        {/* Endpoint Header */}
        <EndpointHeader
          method="POST"
          path="api/auth/register"
          statusCode={201}
          responseTime="260ms"
          responseSize="0.16KB"
        />

        {/* Request/Response Grid */}
        <div className="flex-1 grid grid-cols-2 overflow-hidden">
          {/* Request Panel */}
          <RequestPanel 
            requestBody={requestBody} 
            onExecute={handleExecute} 
          />

          {/* Response Panel */}
          <ResponsePanel responseData={responseData} />
        </div>
      </div>
    </div>
  )
}
