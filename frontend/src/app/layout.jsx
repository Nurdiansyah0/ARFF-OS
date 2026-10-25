import './globals.css';

export const metadata = {
  title: 'ARFF-OS — Airport Rescue & Firefighting Operating System',
  description: 'Enterprise Airport Rescue and Firefighting Operating System for Hang Nadim International Airport',
  icons: {
    icon: '/Logo_hangnadim.png',
  },
};

export default function RootLayout({ children }) {
  return (
    <html lang="id">
      <body className="antialiased min-h-screen bg-slate-50 text-slate-800">
        {children}
      </body>
    </html>
  );
}
