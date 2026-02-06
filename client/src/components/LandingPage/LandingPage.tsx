import { Link } from 'react-router-dom';
import { Baby, Upload, Sparkles, BookOpen, ArrowRight } from 'lucide-react';

export function LandingPage() {
  return (
    <div className="min-h-screen bg-linear-to-br from-pink-50 via-purple-50 to-blue-50">
      {/* Header */}
      <header className="bg-white/80 backdrop-blur-sm border-b border-pink-100">
        <div className="max-w-6xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-linear-to-br from-pink-400 to-purple-500 flex items-center justify-center">
              <Baby className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-xl font-bold bg-linear-to-r from-pink-600 to-purple-600 bg-clip-text text-transparent">
                BabySteps AI Journal
              </h1>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <Link
              to="/sign-in"
              className="px-4 py-2 text-gray-600 hover:text-pink-600 font-medium text-sm transition-colors"
            >
              Sign In
            </Link>
            <Link
              to="/sign-up"
              className="px-5 py-2 bg-linear-to-r from-pink-500 to-purple-500 text-white rounded-xl font-medium text-sm shadow-lg shadow-pink-500/25 hover:shadow-pink-500/40 transition-all"
            >
              Get Started
            </Link>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="max-w-6xl mx-auto px-6 py-20 text-center">
        <div className="inline-flex items-center gap-2 px-4 py-2 bg-pink-100 rounded-full text-pink-600 text-sm font-medium mb-6">
          <Sparkles className="w-4 h-4" />
          AI-Powered Memory Book
        </div>
        
        <h2 className="text-5xl font-bold text-gray-800 mb-6 leading-tight">
          Turn Your Baby Photos Into
          <span className="block bg-linear-to-r from-pink-600 to-purple-600 bg-clip-text text-transparent">
            Magical Memory Books
          </span>
        </h2>
        
        <p className="text-xl text-gray-600 max-w-2xl mx-auto mb-10">
          Upload your photos and let AI organize them into beautiful, personalized 
          journal pages with heartfelt captions and themes.
        </p>

        <div className="flex items-center justify-center gap-4">
          <Link
            to="/sign-up"
            className="px-8 py-4 bg-linear-to-r from-pink-500 to-purple-500 text-white rounded-2xl font-semibold text-lg shadow-xl shadow-pink-500/25 hover:shadow-pink-500/40 hover:scale-105 transition-all flex items-center gap-2"
          >
            Start Your Journal
            <ArrowRight className="w-5 h-5" />
          </Link>
          <Link
            to="/sign-in"
            className="px-8 py-4 bg-white text-gray-700 rounded-2xl font-semibold text-lg shadow-lg hover:shadow-xl transition-all border border-gray-200"
          >
            Sign In
          </Link>
        </div>
      </section>

      {/* Features Section */}
      <section className="max-w-6xl mx-auto px-6 py-16">
        <div className="grid md:grid-cols-3 gap-8">
          <div className="bg-white rounded-3xl p-8 shadow-lg border border-pink-100">
            <div className="w-14 h-14 rounded-2xl bg-linear-to-br from-pink-100 to-pink-200 flex items-center justify-center mb-6">
              <Upload className="w-7 h-7 text-pink-600" />
            </div>
            <h3 className="text-xl font-bold text-gray-800 mb-3">Upload Photos</h3>
            <p className="text-gray-600">
              Simply upload your favorite baby photos. Our AI handles the rest, 
              organizing them by date and moments.
            </p>
          </div>

          <div className="bg-white rounded-3xl p-8 shadow-lg border border-purple-100">
            <div className="w-14 h-14 rounded-2xl bg-linear-to-br from-purple-100 to-purple-200 flex items-center justify-center mb-6">
              <Sparkles className="w-7 h-7 text-purple-600" />
            </div>
            <h3 className="text-xl font-bold text-gray-800 mb-3">AI Magic</h3>
            <p className="text-gray-600">
              Our AI analyzes your photos, suggests themes, and writes heartfelt 
              captions that capture each precious moment.
            </p>
          </div>

          <div className="bg-white rounded-3xl p-8 shadow-lg border border-blue-100">
            <div className="w-14 h-14 rounded-2xl bg-linear-to-br from-blue-100 to-blue-200 flex items-center justify-center mb-6">
              <BookOpen className="w-7 h-7 text-blue-600" />
            </div>
            <h3 className="text-xl font-bold text-gray-800 mb-3">Beautiful Book</h3>
            <p className="text-gray-600">
              Review and customize your pages, then enjoy a beautiful digital 
              memory book you'll treasure forever.
            </p>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="max-w-4xl mx-auto px-6 py-16 text-center">
        <div className="bg-linear-to-r from-pink-500 to-purple-500 rounded-3xl p-12 shadow-2xl">
          <h3 className="text-3xl font-bold text-white mb-4">
            Start Preserving Memories Today
          </h3>
          <p className="text-pink-100 mb-8 text-lg">
            Join thousands of parents creating beautiful memory books for their little ones.
          </p>
          <Link
            to="/sign-up"
            className="inline-flex items-center gap-2 px-8 py-4 bg-white text-pink-600 rounded-2xl font-semibold text-lg shadow-lg hover:shadow-xl hover:scale-105 transition-all"
          >
            Create Free Account
            <ArrowRight className="w-5 h-5" />
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="text-center py-8 text-gray-400 text-sm">
        Made with 💕 for your little ones
      </footer>
    </div>
  );
}
