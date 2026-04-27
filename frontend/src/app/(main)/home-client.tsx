'use client'

import { useState } from 'react'
import Header from '@/components/page/Header'
import Footer from '@/components/page/Footer'
import LoginForm from '@/components/auth/login-form'

export default function HomeClient() {
  const [showAuthModal, setShowAuthModal] = useState(false)

  const handlePageClick = () => {
    setShowAuthModal(true)
  }

  return (
    <div className="flex flex-col min-h-screen">
      <div className={`flex flex-col flex-grow ${showAuthModal ? 'blur-sm' : ''}`}>
        <Header />
        <main className="flex-grow cursor-pointer flex items-center justify-center bg-gray-50" onClick={handlePageClick}>
        </main>
        <Footer />
      </div>
      
      {showAuthModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-8 max-w-sm w-full shadow-xl">
            <button
              onClick={() => setShowAuthModal(false)}
              className="float-right text-gray-600 hover:text-gray-800 text-2xl"
            >
              ×
            </button>
            <LoginForm onSuccess={() => setShowAuthModal(false)} />
          </div>
        </div>
      )}
    </div>
  )
}
