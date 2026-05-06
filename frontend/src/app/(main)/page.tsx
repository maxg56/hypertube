import type { Metadata } from 'next'
import Thumbnails from '@/components/page/Thumbnails'

export const metadata: Metadata = {
  title: 'Accueil — Hypertube',
}

export default function MainPage() {
  // const [showAuthModal, setShowAuthModal] = useState(false)

  // const handlePageClick = () => {
  //   setShowAuthModal(true)
  // }
  return (
      // <div className="flex flex-col min-h-screen">
        <main 
          className="flex-grow overflow-y-auto mt-16 mb-16 cursor-pointer flex items-center justify-center bg-gray-50" 
          // onClick={handlePageClick}
        >
          <Thumbnails />
        </main>
  
        
        /*{/* {showAuthModal && (
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
        )} */
      // </div>
  )
}