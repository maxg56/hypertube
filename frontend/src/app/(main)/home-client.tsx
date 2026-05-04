'use client'

import { useState } from 'react'
import Header from '@/components/page/Header'
import Footer from '@/components/page/Footer'
import LoginForm from '@/components/auth/login-form'
import Thumbnails from '@/components/page/Thumbnails'

export default function HomeClient() {
<<<<<<< 24-frontend-bibliothèque-de-films-avec-infinite-scroll
  // const [showAuthModal, setShowAuthModal] = useState(false)

  // const handlePageClick = () => {
  //   setShowAuthModal(true)
  // }
  
  return (
    <div className="flex flex-col min-h-screen">
      <div className={`fixed top-0 left-0 right-0 z-40`}> 
        {/* /*${/*showAuthModal ? 'blur-sm' /*:  ''}`}`}>*/ }
=======
  const [showAuthModal, setShowAuthModal] = useState(false)

  const handlePageClick = () => {
    setShowAuthModal(true)
  }

  return (
    <div className="flex flex-col min-h-screen">
      <div className={`fixed top-0 left-0 right-0 z-40 ${showAuthModal ? 'blur-sm' : ''}`}>
>>>>>>> main
        <Header />
      </div>

      <main 
        className="flex-grow overflow-y-auto mt-16 mb-16 cursor-pointer flex items-center justify-center bg-gray-50" 
<<<<<<< 24-frontend-bibliothèque-de-films-avec-infinite-scroll
        // onClick={handlePageClick}
=======
        onClick={handlePageClick}
>>>>>>> main
      >
        <Thumbnails />
      </main>

<<<<<<< 24-frontend-bibliothèque-de-films-avec-infinite-scroll
      <div className={`fixed bottom-0 left-0 right-0 z-40`}> 
        {/* // $showAuthModal ?'blur-sm' /*:  ''}``}> */}
        <Footer />
      </div>
      
      {/* {showAuthModal && (
=======
      <div className={`fixed bottom-0 left-0 right-0 z-40 ${showAuthModal ? 'blur-sm' : ''}`}>
        <Footer />
      </div>
      
      {showAuthModal && (
>>>>>>> main
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
<<<<<<< 24-frontend-bibliothèque-de-films-avec-infinite-scroll
      )} */}
=======
      )}
>>>>>>> main
    </div>
  )
}
