import { SignUp } from '@clerk/clerk-react';
import { Baby } from 'lucide-react';

export function SignUpPage() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-pink-50 via-purple-50 to-blue-50">
      {/* Logo */}
      <div className="mb-8 text-center">
        <div className="w-16 h-16 mx-auto rounded-full bg-gradient-to-br from-pink-400 to-purple-500 flex items-center justify-center mb-4">
          <Baby className="w-10 h-10 text-white" />
        </div>
        <h1 className="text-3xl font-bold bg-gradient-to-r from-pink-600 to-purple-600 bg-clip-text text-transparent">
          BabySteps AI Journal
        </h1>
        <p className="text-gray-500 mt-2">Create your account</p>
      </div>

      {/* Clerk Sign Up Component */}
      <SignUp
        appearance={{
          elements: {
            rootBox: 'mx-auto',
            card: 'shadow-xl border border-pink-100',
            headerTitle: 'text-gray-800',
            headerSubtitle: 'text-gray-500',
            socialButtonsBlockButton: 'border-gray-200 hover:bg-pink-50',
            formButtonPrimary: 'bg-gradient-to-r from-pink-500 to-purple-500 hover:from-pink-600 hover:to-purple-600',
            footerActionLink: 'text-pink-600 hover:text-pink-700',
          },
        }}
        routing="path"
        path="/sign-up"
        signInUrl="/sign-in"
      />
    </div>
  );
}
