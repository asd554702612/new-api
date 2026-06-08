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
import { Shield, Sparkles, Users } from 'lucide-react';

const iconMap = {
  shield: Shield,
  sparkles: Sparkles,
  users: Users,
};

const HomeShowcase = (props) => {
  return (
    <section className='mx-auto mt-10 max-w-7xl'>
      <div className='mb-10 text-center'>
        <p className='text-sm font-black uppercase tracking-[0.22em] text-slate-500 dark:text-slate-400'>
          {props.data.kicker}
        </p>
        <h2 className='mt-3 text-3xl font-black tracking-tight text-slate-950 dark:text-white'>
          {props.data.title}
        </h2>
        <p className='mx-auto mt-4 max-w-3xl text-sm leading-7 text-slate-500 dark:text-slate-300 sm:text-base'>
          {props.data.subtitle}
        </p>
      </div>

      <div className='grid gap-5 lg:grid-cols-3'>
        {props.data.stats.map((stat) => {
          const Icon = iconMap[stat.icon];
          return (
            <article
              key={stat.label}
              className='rounded-[2rem] border border-slate-200/80 bg-white/[0.92] p-6 text-center shadow-[0_16px_50px_rgba(148,163,184,0.12)] dark:border-slate-800/80 dark:bg-slate-900/[0.78]'
            >
              <div className='mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-slate-950 text-white dark:bg-cyan-300 dark:text-slate-950'>
                <Icon size={28} />
              </div>
              <p className='mt-5 text-4xl font-black text-slate-950 dark:text-white'>
                {stat.value}
              </p>
              <p className='mt-2 text-sm font-medium text-slate-500 dark:text-slate-300'>
                {stat.label}
              </p>
            </article>
          );
        })}
      </div>

      <div className='mt-5 grid gap-5 md:grid-cols-3'>
        {props.data.quotes.map((quote) => (
          <article
            key={quote.author}
            className='rounded-[2rem] border border-slate-200/80 bg-white/[0.92] p-6 shadow-[0_16px_50px_rgba(148,163,184,0.12)] dark:border-slate-800/80 dark:bg-slate-900/[0.78]'
          >
            <div className='flex gap-1'>
              {[1, 2, 3, 4, 5].map((star) => (
                <span key={star} className='text-yellow-500'>
                  ★
                </span>
              ))}
            </div>
            <p className='mt-4 text-sm leading-7 text-slate-600 dark:text-slate-300'>
              “{quote.quote}”
            </p>
            <div className='mt-5'>
              <p className='text-sm font-black text-slate-950 dark:text-white'>
                {quote.author}
              </p>
              <p className='mt-1 text-xs text-slate-400 dark:text-slate-500'>
                {quote.role}
              </p>
            </div>
          </article>
        ))}
      </div>
    </section>
  );
};

export default HomeShowcase;
