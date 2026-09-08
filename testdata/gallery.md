# termtex gallery

This document exercises the markdown path: every equation from demo.sh as display math, plus inline uses.

Inline: the quadratic $x = \frac{-b \pm \sqrt{b^2-4ac}}{2a}$, a sum $\sum_{i=1}^{n} i^2$, an integral $\int_0^1 f(x)\,dx$, a limit $\lim_{n\to\infty} a_n$, and a norm $\|x\|_2 \le \|x\|_1$ all stay in the sentence. Prices like $5 and $10,000 are not math.

**1. Euler's Identity**

$$e^{i\pi} + 1 = 0$$

**2. Pythagorean Theorem**

$$a^2 + b^2 = c^2$$

**3. Quadratic Formula**

$$\frac{-b \pm \sqrt{b^2 - 4ac}}{2a}$$

**4. Binomial Theorem**

$$(x + y)^n = \sum_{k=0}^{n} \frac{n!}{k!(n-k)!} x^k y^{n-k}$$

**5. Power Rule**

$$\frac{d}{dx} x^n = nx^{n-1}$$

**6. Definition of Derivative**

$$\frac{df}{dx} = \lim_{h \to 0} \frac{f(x+h) - f(x)}{h}$$

**7. Fundamental Theorem of Calculus**

$$\int_{a}^{b} f(x) dx = F(b) - F(a)$$

**8. Chain Rule**

$$\frac{dy}{dx} = \frac{dy}{du} \cdot \frac{du}{dx}$$

**9. Product Rule**

$$(fg)^{\prime} = f^{\prime}g + fg^{\prime}$$

**10. Integration by Parts**

$$\int u \, dv = uv - \int v \, du$$

**11. Sum of Natural Numbers**

$$\sum_{i=1}^{n} i = \frac{n(n+1)}{2}$$

**12. Sum of Squares**

$$\sum_{i=1}^{n} i^2 = \frac{n(n+1)(2n+1)}{6}$$

**13. Geometric Series**

$$\sum_{k=0}^{n} r^k = \frac{1 - r^{n+1}}{1 - r}$$

**14. Infinite Geometric Series**

$$\sum_{k=0}^{\infty} r^k = \frac{1}{1 - r}$$

**15. Difference of Squares**

$$a^2 - b^2 = (a + b)(a - b)$$

**16. Cubic Formula (Depressed)**

$$x = \sqrt[3]{-\frac{q}{2} + \sqrt{\frac{q^2}{4} + \frac{p^3}{27}}} + \sqrt[3]{-\frac{q}{2} - \sqrt{\frac{q^2}{4} + \frac{p^3}{27}}}$$

**17. Fraction Addition**

$$\frac{a}{b} + \frac{c}{d} = \frac{ad + bc}{bd}$$

**18. Logarithm Change of Base**

$$\log_a b = \frac{\ln b}{\ln a}$$

**19. Exponential-Log Inverse**

$$e^{\ln x} = x$$

**20. Euler's Totient Product**

$$\phi(n) = n \prod_{p | n} \left(1 - \frac{1}{p}\right)$$

**21. Gaussian Integral**

$$\int_{0}^{\infty} e^{-x^2} dx = \frac{\sqrt{\pi}}{2}$$

**22. Taylor Series**

$$f(x) = \sum_{n=0}^{\infty} \frac{f^{(n)}(a)}{n!} (x - a)^n$$

**23. Maclaurin Series for e^x**

$$e^x = \sum_{n=0}^{\infty} \frac{x^n}{n!}$$

**24. Basel Problem**

$$\sum_{n=1}^{\infty} \frac{1}{n^2} = \frac{\pi^2}{6}$$

**25. Leibniz Formula for Pi**

$$\frac{\pi}{4} = 1 - \frac{1}{3} + \frac{1}{5} - \frac{1}{7} + \cdots$$

**26. Cauchy-Schwarz Inequality**

$$\left(\sum a_i b_i\right)^2 \leq \left(\sum a_i^2\right)\left(\sum b_i^2\right)$$

**27. Mean Value Theorem**

$$f(b) - f(a) = f^{\prime}(c)(b - a)$$

**28. L'Hopital's Rule**

$$\lim_{x \to c} \frac{f(x)}{g(x)} = \lim_{x \to c} \frac{f^{\prime}(x)}{g^{\prime}(x)}$$

**29. Euler-Mascheroni Constant**

$$\gamma = \lim_{n \to \infty} \left(\sum_{k=1}^{n} \frac{1}{k} - \ln n\right)$$

**30. Stirling's Approximation**

$$n! \approx \sqrt{2\pi n} \left(\frac{n}{e}\right)^n$$

**31. Pythagorean Identity**

$$\sin^2 \theta + \cos^2 \theta = 1$$

**32. Sine Addition**

$$\sin(\alpha + \beta) = \sin \alpha \cos \beta + \cos \alpha \sin \beta$$

**33. Cosine Addition**

$$\cos(\alpha + \beta) = \cos \alpha \cos \beta - \sin \alpha \sin \beta$$

**34. Euler's Formula**

$$e^{i\theta} = \cos \theta + i \sin \theta$$

**35. Double Angle Sine**

$$\sin 2\theta = 2 \sin \theta \cos \theta$$

**36. Double Angle Cosine**

$$\cos 2\theta = \cos^2 \theta - \sin^2 \theta$$

**37. Tangent Definition**

$$\tan \theta = \frac{\sin \theta}{\cos \theta}$$

**38. Law of Sines**

$$\frac{a}{\sin A} = \frac{b}{\sin B} = \frac{c}{\sin C}$$

**39. Law of Cosines**

$$c^2 = a^2 + b^2 - 2ab \cos C$$

**40. Half-Angle Formula**

$$\sin \frac{\theta}{2} = \pm \sqrt{\frac{1 - \cos \theta}{2}}$$

**41. 2x2 Determinant**

$$\det \begin{bmatrix} a & b \\ c & d \end{bmatrix} = ad - bc$$

**42. Kronecker Delta**

$$\delta_{ij} = \begin{cases} 1 & \text{if } i = j \\ 0 & \text{if } i \neq j \end{cases}$$

**43. Matrix Inverse (2x2)**

$$A^{-1} = \frac{1}{ad - bc} \begin{bmatrix} d & -b \\ -c & a \end{bmatrix}$$

**44. Eigenvalue Equation**

$$Av = \lambda v$$

**45. Characteristic Polynomial**

$$\det(A - \lambda I) = 0$$

**46. Dot Product**

$$a \cdot b = \sum_{i=1}^{n} a_i b_i$$

**47. Cross Product**

$$a \times b = \begin{vmatrix} i & j & k \\ a_1 & a_2 & a_3 \\ b_1 & b_2 & b_3 \end{vmatrix}$$

**48. Matrix Transpose**

$$(AB)^T = B^T A^T$$

**49. Trace**

$$\text{tr}(A) = \sum_{i=1}^{n} a_{ii}$$

**50. Rotation Matrix**

$$R(\theta) = \begin{bmatrix} \cos \theta & -\sin \theta \\ \sin \theta & \cos \theta \end{bmatrix}$$

**51. Newton's Second Law**

$$F = ma$$

**52. E = mc^2**

$$E = mc^2$$

**53. Kinetic Energy**

$$E_k = \frac{1}{2}mv^2$$

**54. Gravitational Force**

$$F = G\frac{m_1 m_2}{r^2}$$

**55. Gravitational Potential Energy**

$$U = -\frac{Gm_1 m_2}{r}$$

**56. Escape Velocity**

$$v_e = \sqrt{\frac{2GM}{r}}$$

**57. Simple Harmonic Motion**

$$x(t) = A\cos(\omega t + \phi)$$

**58. Euler-Lagrange Equation**

$$\frac{d}{dt}\frac{\partial L}{\partial v} - \frac{\partial L}{\partial q} = 0$$

**59. Hamilton's Equation**

$$\frac{dq}{dt} = \frac{\partial H}{\partial p}$$

**60. Centripetal Acceleration**

$$a = \frac{v^2}{r}$$

**61. Gauss's Law**

$$\nabla \cdot E = \frac{\rho}{\epsilon_0}$$

**62. Gauss's Law (Magnetism)**

$$\nabla \cdot B = 0$$

**63. Faraday's Law**

$$\nabla \times E = -\frac{\partial B}{\partial t}$$

**64. Ampere's Law**

$$\nabla \times B = \mu_0 J + \mu_0 \epsilon_0 \frac{\partial E}{\partial t}$$

**65. Coulomb's Law**

$$F = \frac{1}{4\pi\epsilon_0} \cdot \frac{q_1 q_2}{r^2}$$

**66. Lorentz Force**

$$F = q(E + v \times B)$$

**67. Ohm's Law**

$$V = IR$$

**68. Capacitor Energy**

$$E = \frac{1}{2}CV^2$$

**69. Biot-Savart Law**

$$dB = \frac{\mu_0}{4\pi} \frac{I \, dl \times \hat{r}}{r^2}$$

**70. Poynting Vector**

$$S = \frac{1}{\mu_0} E \times B$$

**71. Schrodinger Equation**

$$i\hbar\frac{\partial}{\partial t}\Psi = \hat{H}\Psi$$

**72. Heisenberg Uncertainty**

$$\Delta x \, \Delta p \geq \frac{\hbar}{2}$$

**73. de Broglie Wavelength**

$$\lambda = \frac{h}{p}$$

**74. Planck-Einstein Relation**

$$E = h\nu$$

**75. Time Dilation**

$$\Delta t^{\prime} = \frac{\Delta t}{\sqrt{1 - \frac{v^2}{c^2}}}$$

**76. Length Contraction**

$$L = L_0 \sqrt{1 - \frac{v^2}{c^2}}$$

**77. Relativistic Energy-Momentum**

$$E^2 = (pc)^2 + (mc^2)^2$$

**78. Schwarzschild Radius**

$$r_s = \frac{2GM}{c^2}$$

**79. Photoelectric Effect**

$$E_k = h\nu - \phi$$

**80. Rydberg Formula**

$$\frac{1}{\lambda} = R\left(\frac{1}{{n_1}^2} - \frac{1}{{n_2}^2}\right)$$

**81. Ideal Gas Law**

$$PV = nRT$$

**82. Boltzmann Entropy**

$$S = k_B \ln \Omega$$

**83. First Law of Thermodynamics**

$$\Delta U = Q - W$$

**84. Carnot Efficiency**

$$\eta = 1 - \frac{T_c}{T_h}$$

**85. Stefan-Boltzmann Law**

$$P = \sigma A T^4$$

**86. Maxwell-Boltzmann Distribution**

$$f(v) = 4\pi n \left(\frac{m}{2\pi kT}\right)^{3/2} v^2 e^{-mv^2/2kT}$$

**87. Planck's Law**

$$B(\nu) = \frac{2h\nu^3}{c^2} \cdot \frac{1}{e^{h\nu/kT} - 1}$$

**88. Gibbs Free Energy**

$$G = H - TS$$

**89. Clausius Inequality**

$$\oint \frac{dQ}{T} \leq 0$$

**90. Equipartition Theorem**

$$\langle E \rangle = \frac{f}{2} k_B T$$

**91. Bayes' Theorem**

$$P(A \mid B) = \frac{P(B \mid A) \, P(A)}{P(B)}$$

**92. Normal Distribution**

$$f(x) = \frac{1}{\sigma\sqrt{2\pi}} e^{-\frac{(x - \mu)^2}{2\sigma^2}}$$

**93. Expected Value**

$$E[X] = \sum_{i} x_i \, P(x_i)$$

**94. Variance**

$$\text{Var}(X) = E[X^2] - (E[X])^2$$

**95. Shannon Entropy**

$$H = -\sum_{i} p_i \log_2 p_i$$

**96. Bernoulli Trial**

$$P(k) = \frac{n!}{k!(n-k)!} p^k (1-p)^{n-k}$$

**97. Golden Ratio**

$$\phi = \frac{1 + \sqrt{5}}{2}$$

**98. Euler Product (Riemann Zeta)**

$$\zeta(s) = \sum_{n=1}^{\infty} \frac{1}{n^s} = \prod_{p} \frac{1}{1 - p^{-s}}$$

**99. Wallis Product**

$$\frac{\pi}{2} = \prod_{n=1}^{\infty} \frac{4n^2}{4n^2 - 1}$$

**100. Euler's Reflection Formula**

$$\Gamma(z)\Gamma(1-z) = \frac{\pi}{\sin(\pi z)}$$

