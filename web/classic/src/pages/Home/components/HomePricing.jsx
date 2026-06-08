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
import { Link } from 'react-router-dom';

const HomePricing = (props) => {
  return (
    <section
      id='pricing'
      className='mx-auto mt-10 max-w-7xl rounded-[2rem] border border-slate-200/80 bg-white/[0.92] px-6 py-8 shadow-[0_16px_50px_rgba(148,163,184,0.12)] dark:border-slate-800/80 dark:bg-slate-900/[0.78] sm:px-8'
    >
      <div className='flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between'>
        <div>
          <p className='text-sm font-black uppercase tracking-[0.22em] text-slate-500 dark:text-slate-400'>
            {props.data.kicker}
          </p>
          <h2 className='mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white'>
            {props.data.title}
          </h2>
          <p className='mt-3 max-w-3xl text-sm leading-7 text-slate-500 dark:text-slate-300 sm:text-base'>
            {props.data.subtitle}
          </p>
        </div>
        <Link
          to='/pricing'
          className='inline-flex items-center text-sm font-semibold text-blue-600 transition-colors hover:text-blue-700 dark:text-cyan-300 dark:hover:text-cyan-200'
        >
          {props.data.viewDetails} →
        </Link>
      </div>

      <div className='mt-8 grid gap-5 lg:grid-cols-3'>
        {props.data.rows.map((row) => (
          <Link
            key={row.plan}
            to='/pricing'
            className={`rounded-[1.8rem] border px-5 py-5 transition-transform hover:-translate-y-1 ${
              row.featured
                ? 'border-blue-200 bg-gradient-to-br from-blue-50 via-white to-cyan-50 dark:border-cyan-400/20 dark:from-cyan-500/10 dark:via-slate-900/80 dark:to-blue-500/10'
                : 'border-slate-200/80 bg-slate-50/90 dark:border-slate-800/80 dark:bg-slate-950/[0.85]'
            }`}
          >
            <div className='flex items-start justify-between gap-4'>
              <div>
                <div className='flex items-center gap-2'>
                  <p className='text-base font-black text-slate-950 dark:text-white'>
                    {row.plan}
                  </p>
                  <span
                    className={`inline-flex rounded-full px-2.5 py-1 text-[11px] font-black ${
                      row.featured
                        ? 'bg-blue-600 text-white dark:bg-cyan-300 dark:text-slate-950'
                        : 'bg-slate-200 text-slate-600 dark:bg-slate-800 dark:text-slate-100'
                    }`}
                  >
                    {row.badge}
                  </span>
                </div>
                <p className='mt-2 text-sm leading-7 text-slate-500 dark:text-slate-300'>
                  {row.description}
                </p>
              </div>
              <p className='whitespace-nowrap text-xl font-black text-slate-950 dark:text-white'>
                {row.price}
              </p>
            </div>
          </Link>
        ))}
      </div>
    </section>
  );
};

export default HomePricing;
