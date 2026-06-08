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
import { BarChart3, Shield, Shuffle } from 'lucide-react';

const iconMap = {
  chartBar: BarChart3,
  shield: Shield,
  swap: Shuffle,
};

const HomeFeatures = (props) => {
  return (
    <section
      id='overview'
      className='relative overflow-hidden bg-gradient-to-b from-white via-zinc-50/30 to-white px-4 py-32 sm:px-6 lg:px-8 dark:from-[#040816] dark:via-[#071022] dark:to-[#040816]'
    >
      <div className='absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(15,23,42,0.02),transparent_65%)] dark:bg-[radial-gradient(circle_at_center,rgba(255,255,255,0.02),transparent_65%)]' />
      <div className='relative z-10 mx-auto max-w-7xl'>
        <div className='mb-20 text-center'>
          <span className='mb-4 inline-block text-sm font-semibold uppercase tracking-widest text-zinc-500 dark:text-slate-400'>
            {props.data.kicker}
          </span>
          <h2 className='mb-6 font-serif text-5xl leading-tight text-black dark:text-white md:text-6xl'>
            {props.data.title}
          </h2>
          <p className='mx-auto max-w-2xl text-lg text-zinc-600 dark:text-slate-300'>
            {props.data.subtitle}
          </p>
        </div>

        <div className='grid grid-cols-1 gap-6 md:grid-cols-3'>
          {props.data.cards.map((card, index) => {
            const Icon = iconMap[card.icon];
            const cardClassName =
              index === 0
                ? 'md:col-span-2 border-zinc-100 bg-zinc-50 hover:border-zinc-200 dark:border-slate-800 dark:bg-slate-900/70 dark:hover:border-slate-700'
                : index === 1
                  ? 'border-zinc-100 bg-white hover:border-zinc-200 dark:border-slate-800 dark:bg-slate-900/70 dark:hover:border-slate-700'
                  : 'border-zinc-100 bg-zinc-900 text-white hover:border-zinc-800 dark:border-slate-800';

            return (
              <article
                key={card.title}
                className={`group relative overflow-hidden rounded-[2.5rem] border p-8 transition-all duration-500 hover:shadow-2xl ${cardClassName}`}
              >
                <div
                  className={`relative z-10 ${index === 0 ? 'max-w-md' : ''}`}
                >
                  <div
                    className={`mb-6 flex h-14 w-14 items-center justify-center rounded-2xl ${
                      index === 0
                        ? 'bg-black text-white dark:bg-cyan-300 dark:text-slate-950'
                        : index === 1
                          ? 'bg-zinc-100 text-black dark:bg-slate-800 dark:text-white'
                          : 'bg-white/10 text-white'
                    }`}
                  >
                    <Icon size={28} />
                  </div>
                  <h3
                    className={`mb-4 text-3xl font-bold leading-tight ${
                      index === 2 ? 'text-white' : 'text-black dark:text-white'
                    }`}
                  >
                    {card.title}
                  </h3>
                  <p
                    className={`text-sm leading-relaxed ${
                      index === 2
                        ? 'text-zinc-300'
                        : 'text-zinc-600 dark:text-slate-300'
                    }`}
                  >
                    {card.description}
                  </p>
                </div>

                {index === 0 ? (
                  <div className='pointer-events-none absolute right-[-10%] top-10 h-[300px] w-[300px] rounded-full bg-blue-500/5 blur-[80px] dark:bg-cyan-400/10' />
                ) : null}
                {index === 1 ? (
                  <div className='absolute bottom-0 left-0 right-0 p-8'>
                    <div className='flex flex-col gap-2'>
                      {[45, 72, 88].map((bar) => (
                        <div
                          key={bar}
                          className='h-2 overflow-hidden rounded-full bg-zinc-50 dark:bg-slate-800'
                        >
                          <div
                            className='h-full bg-zinc-200 dark:bg-slate-600'
                            style={{ width: `${bar}%` }}
                          />
                        </div>
                      ))}
                    </div>
                  </div>
                ) : null}
                {index === 2 ? (
                  <div className='mt-8 rounded-2xl border border-white/10 bg-white/5 p-6 font-mono text-xs text-zinc-300'>
                    <code className='block leading-7'>
                      import International from "international"
                      <br />
                      const client = new International()
                      <br />
                      await client.responses.create()
                    </code>
                  </div>
                ) : null}
              </article>
            );
          })}
        </div>
      </div>
    </section>
  );
};

export default HomeFeatures;
