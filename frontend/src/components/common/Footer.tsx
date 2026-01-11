import { Link } from 'react-router-dom'

export default function Footer() {
  return (
    <footer className="bg-carbon-dark border-t border-steel">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex flex-col md:flex-row items-center justify-between gap-4">
          {/* Logo & Copyright */}
          <div className="flex flex-col items-center md:items-start gap-2">
            <Link
              to="/admin"
              className="text-xl font-heading font-bold text-white tracking-tight hover:opacity-80 transition-opacity"
              title=""
            >
              F<span className="text-racing">R</span>C
            </Link>
            <p className="text-text-muted text-sm">
              © 2024 F1 Rivals Cup. All rights reserved.
            </p>
          </div>

          {/* Contact Info */}
          <div className="flex flex-col items-center md:items-end gap-2">
            <p className="text-text-secondary text-sm">
              Contact Us
            </p>
            <a
              href="mailto:kangmomin@inab.kr"
              className="text-neon hover:text-neon-light transition-colors duration-150"
            >
              kangmomin@inab.kr
            </a>
          </div>
        </div>

        {/* Bottom Links */}
        <div className="mt-8 pt-4 border-t border-steel flex flex-wrap justify-center gap-6 text-sm text-text-muted">
          <a href="#" className="hover:text-text-secondary transition-colors">이용약관</a>
          <a href="#" className="hover:text-text-secondary transition-colors">개인정보처리방침</a>
          <a href="#" className="hover:text-text-secondary transition-colors">문의하기</a>
        </div>
      </div>
    </footer>
  )
}
