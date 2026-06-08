/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React from 'react';

const HomeFooter = (props) => {
  const supportContactLines = String(props.supportContactInfo || '')
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);

  return (
    <footer
      id='footer'
      className='relative z-10 mt-10 bg-[#0a1020] px-4 pb-8 pt-8 text-white sm:px-6'
    >
      <div className='mx-auto max-w-7xl rounded-[28px] border border-white/10 bg-slate-950/45 px-6 py-8 backdrop-blur-xl sm:px-8'>
        <div className='grid gap-8 lg:grid-cols-[1.1fr_1.4fr_0.55fr]'>
          <div>
            <div className='flex items-center gap-3'>
              <div className='flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-white/10'>
                <img
                  src={props.siteLogo || '/favicon.ico'}
                  alt={props.brand}
                  className='h-full w-full object-contain p-1.5'
                />
              </div>
              <div>
                <p className='text-lg font-black'>{props.brand}</p>
                <p className='text-sm text-slate-300'>{props.data.about}</p>
              </div>
            </div>

            <div className='mt-5 flex gap-3 text-slate-400'>
              {props.data.socialDots.map((dot) => (
                <span
                  key={dot}
                  className='inline-flex h-9 w-9 items-center justify-center rounded-full border border-white/10 bg-white/5 text-sm dark:border-cyan-400/10 dark:bg-slate-900/70'
                >
                  {dot}
                </span>
              ))}
            </div>
          </div>

          <div className='grid gap-8 sm:grid-cols-4'>
            {props.data.columns.map((column) => (
              <div key={column.title}>
                <p className='text-sm font-black text-white'>{column.title}</p>
                <div className='mt-4 space-y-3'>
                  {column.items.map((item) => (
                    <a
                      key={`${column.title}-${item.label}`}
                      href={item.href}
                      target={item.external ? '_blank' : undefined}
                      rel={item.external ? 'noopener noreferrer' : undefined}
                      className='block text-sm text-slate-300 transition-colors hover:text-white dark:hover:text-cyan-200'
                    >
                      {item.label}
                    </a>
                  ))}
                </div>
              </div>
            ))}
          </div>

          {supportContactLines.length > 0 ? (
            <div>
              <p className='text-sm font-black text-white'>
                {props.supportContactLabel}
              </p>
              <div className='mt-4 space-y-3'>
                {supportContactLines.map((line, index) => (
                  <p
                    key={`${line}-${index}`}
                    className='break-words text-sm leading-6 text-slate-300'
                  >
                    {line}
                  </p>
                ))}
              </div>
            </div>
          ) : null}
        </div>

        <div className='mt-8 flex flex-col items-center justify-center gap-3 border-t border-white/10 pt-6 text-center text-sm text-slate-300'>
          <a
            href={props.docUrl}
            target='_blank'
            rel='noopener noreferrer'
            className='font-semibold transition-colors hover:text-white'
          >
            {props.docsLabel}
          </a>
        </div>
      </div>
    </footer>
  );
};

export default HomeFooter;
