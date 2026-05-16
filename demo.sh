#!/bin/zsh

C="--color"
SCRIPT_DIR="${0:a:h}"
GALLERY=""

TERMTEX=$(mktemp -t termtex.XXXXXX)
trap 'rm -f "$TERMTEX"' EXIT
go build -o "$TERMTEX" "$SCRIPT_DIR/cmd/termtex" || exit 1

echo ""
echo "\033[38;2;130;180;255m════════════════════════════════════════════════════════════\033[0m"
echo "\033[38;2;130;180;255m  termtex\033[38;2;90;104;137m — Terminal Math Rendering\033[0m"
echo "\033[38;2;130;180;255m════════════════════════════════════════════════════════════\033[0m"

show() {
  echo ""
  echo "\033[38;2;160;210;110m  $1\033[0m"
  printf '  \033[38;2;90;104;137m%s\033[0m\n' "$2"
  echo ""
  local rendered=$(printf '%s\n' "$2" | "$TERMTEX" 2>/dev/null)
  printf '%s\n' "$rendered" | sed 's/^/  /'
  # Collect for README gallery
  GALLERY+="**$1** \`$2\`"$'\n'
  GALLERY+='```'$'\n'
  GALLERY+="$rendered"$'\n'
  GALLERY+='```'$'\n\n'
}

section() {
  echo ""
  echo "\033[38;2;130;180;255m  ── $1 ──\033[0m"
  GALLERY+="### $1"$'\n\n'
}

# ── Foundations ──────────────────────────────────────────────
section "Foundations"

show "1. Euler's Identity" 'e^{i\pi} + 1 = 0'
show "2. Pythagorean Theorem" 'a^2 + b^2 = c^2'
show "3. Quadratic Formula" '\frac{-b \pm \sqrt{b^2 - 4ac}}{2a}'
show "4. Binomial Theorem" '(x + y)^n = \sum_{k=0}^{n} \frac{n!}{k!(n-k)!} x^k y^{n-k}'
show "5. Power Rule" '\frac{d}{dx} x^n = nx^{n-1}'
show "6. Definition of Derivative" '\frac{df}{dx} = \lim_{h \to 0} \frac{f(x+h) - f(x)}{h}'
show "7. Fundamental Theorem of Calculus" '\int_{a}^{b} f(x) dx = F(b) - F(a)'
show "8. Chain Rule" '\frac{dy}{dx} = \frac{dy}{du} \cdot \frac{du}{dx}'
show "9. Product Rule" '(fg)^{\prime} = f^{\prime}g + fg^{\prime}'
show "10. Integration by Parts" '\int u \, dv = uv - \int v \, du'

# ── Algebra & Number Theory ─────────────────────────────────
section "Algebra and Number Theory"

show "11. Sum of Natural Numbers" '\sum_{i=1}^{n} i = \frac{n(n+1)}{2}'
show "12. Sum of Squares" '\sum_{i=1}^{n} i^2 = \frac{n(n+1)(2n+1)}{6}'
show "13. Geometric Series" '\sum_{k=0}^{n} r^k = \frac{1 - r^{n+1}}{1 - r}'
show "14. Infinite Geometric Series" '\sum_{k=0}^{\infty} r^k = \frac{1}{1 - r}'
show "15. Difference of Squares" 'a^2 - b^2 = (a + b)(a - b)'
show "16. Cubic Formula (Depressed)" 'x = \sqrt[3]{-\frac{q}{2} + \sqrt{\frac{q^2}{4} + \frac{p^3}{27}}} + \sqrt[3]{-\frac{q}{2} - \sqrt{\frac{q^2}{4} + \frac{p^3}{27}}}'
show "17. Fraction Addition" '\frac{a}{b} + \frac{c}{d} = \frac{ad + bc}{bd}'
show "18. Logarithm Change of Base" '\log_a b = \frac{\ln b}{\ln a}'
show "19. Exponential-Log Inverse" 'e^{\ln x} = x'
show "20. Euler's Totient Product" '\phi(n) = n \prod_{p | n} \left(1 - \frac{1}{p}\right)'

# ── Calculus & Analysis ─────────────────────────────────────
section "Calculus and Analysis"

show "21. Gaussian Integral" '\int_{0}^{\infty} e^{-x^2} dx = \frac{\sqrt{\pi}}{2}'
show "22. Taylor Series" 'f(x) = \sum_{n=0}^{\infty} \frac{f^{(n)}(a)}{n!} (x - a)^n'
show "23. Maclaurin Series for e^x" 'e^x = \sum_{n=0}^{\infty} \frac{x^n}{n!}'
show "24. Basel Problem" '\sum_{n=1}^{\infty} \frac{1}{n^2} = \frac{\pi^2}{6}'
show "25. Leibniz Formula for Pi" '\frac{\pi}{4} = 1 - \frac{1}{3} + \frac{1}{5} - \frac{1}{7} + \cdots'
show "26. Cauchy-Schwarz Inequality" '\left(\sum a_i b_i\right)^2 \leq \left(\sum a_i^2\right)\left(\sum b_i^2\right)'
show "27. Mean Value Theorem" 'f(b) - f(a) = f^{\prime}(c)(b - a)'
show "28. L'Hopital's Rule" '\lim_{x \to c} \frac{f(x)}{g(x)} = \lim_{x \to c} \frac{f^{\prime}(x)}{g^{\prime}(x)}'
show "29. Euler-Mascheroni Constant" '\gamma = \lim_{n \to \infty} \left(\sum_{k=1}^{n} \frac{1}{k} - \ln n\right)'
show "30. Stirling's Approximation" 'n! \approx \sqrt{2\pi n} \left(\frac{n}{e}\right)^n'

# ── Trigonometry ────────────────────────────────────────────
section "Trigonometry"

show "31. Pythagorean Identity" '\sin^2 \theta + \cos^2 \theta = 1'
show "32. Sine Addition" '\sin(\alpha + \beta) = \sin \alpha \cos \beta + \cos \alpha \sin \beta'
show "33. Cosine Addition" '\cos(\alpha + \beta) = \cos \alpha \cos \beta - \sin \alpha \sin \beta'
show "34. Euler's Formula" 'e^{i\theta} = \cos \theta + i \sin \theta'
show "35. Double Angle Sine" '\sin 2\theta = 2 \sin \theta \cos \theta'
show "36. Double Angle Cosine" '\cos 2\theta = \cos^2 \theta - \sin^2 \theta'
show "37. Tangent Definition" '\tan \theta = \frac{\sin \theta}{\cos \theta}'
show "38. Law of Sines" '\frac{a}{\sin A} = \frac{b}{\sin B} = \frac{c}{\sin C}'
show "39. Law of Cosines" 'c^2 = a^2 + b^2 - 2ab \cos C'
show "40. Half-Angle Formula" '\sin \frac{\theta}{2} = \pm \sqrt{\frac{1 - \cos \theta}{2}}'

# ── Linear Algebra ──────────────────────────────────────────
section "Linear Algebra"

show "41. 2x2 Determinant" '\det \begin{bmatrix} a & b \\ c & d \end{bmatrix} = ad - bc'
show "42. Kronecker Delta" '\delta_{ij} = \begin{cases} 1 & \text{if } i = j \\ 0 & \text{if } i \neq j \end{cases}'
show "43. Matrix Inverse (2x2)" 'A^{-1} = \frac{1}{ad - bc} \begin{bmatrix} d & -b \\ -c & a \end{bmatrix}'
show "44. Eigenvalue Equation" 'Av = \lambda v'
show "45. Characteristic Polynomial" '\det(A - \lambda I) = 0'
show "46. Dot Product" 'a \cdot b = \sum_{i=1}^{n} a_i b_i'
show "47. Cross Product" 'a \times b = \begin{vmatrix} i & j & k \\ a_1 & a_2 & a_3 \\ b_1 & b_2 & b_3 \end{vmatrix}'
show "48. Matrix Transpose" '(AB)^T = B^T A^T'
show "49. Trace" '\text{tr}(A) = \sum_{i=1}^{n} a_{ii}'
show "50. Rotation Matrix" 'R(\theta) = \begin{bmatrix} \cos \theta & -\sin \theta \\ \sin \theta & \cos \theta \end{bmatrix}'

# ── Physics: Classical Mechanics ────────────────────────────
section "Classical Mechanics"

show "51. Newton's Second Law" 'F = ma'
show "52. E = mc^2" 'E = mc^2'
show "53. Kinetic Energy" 'E_k = \frac{1}{2}mv^2'
show "54. Gravitational Force" 'F = G\frac{m_1 m_2}{r^2}'
show "55. Gravitational Potential Energy" 'U = -\frac{Gm_1 m_2}{r}'
show "56. Escape Velocity" 'v_e = \sqrt{\frac{2GM}{r}}'
show "57. Simple Harmonic Motion" 'x(t) = A\cos(\omega t + \phi)'
show "58. Euler-Lagrange Equation" '\frac{d}{dt}\frac{\partial L}{\partial v} - \frac{\partial L}{\partial q} = 0'
show "59. Hamilton's Equation" '\frac{dq}{dt} = \frac{\partial H}{\partial p}'
show "60. Centripetal Acceleration" 'a = \frac{v^2}{r}'

# ── Physics: Electromagnetism ───────────────────────────────
section "Electromagnetism"

show "61. Gauss's Law" '\nabla \cdot E = \frac{\rho}{\epsilon_0}'
show "62. Gauss's Law (Magnetism)" '\nabla \cdot B = 0'
show "63. Faraday's Law" '\nabla \times E = -\frac{\partial B}{\partial t}'
show "64. Ampere's Law" '\nabla \times B = \mu_0 J + \mu_0 \epsilon_0 \frac{\partial E}{\partial t}'
show "65. Coulomb's Law" 'F = \frac{1}{4\pi\epsilon_0} \cdot \frac{q_1 q_2}{r^2}'
show "66. Lorentz Force" 'F = q(E + v \times B)'
show "67. Ohm's Law" 'V = IR'
show "68. Capacitor Energy" 'E = \frac{1}{2}CV^2'
show "69. Biot-Savart Law" 'dB = \frac{\mu_0}{4\pi} \frac{I \, dl \times \hat{r}}{r^2}'
show "70. Poynting Vector" 'S = \frac{1}{\mu_0} E \times B'

# ── Physics: Quantum & Relativity ──────────────────────────
section "Quantum Mechanics and Relativity"

show "71. Schrodinger Equation" 'i\hbar\frac{\partial}{\partial t}\Psi = \hat{H}\Psi'
show "72. Heisenberg Uncertainty" '\Delta x \, \Delta p \geq \frac{\hbar}{2}'
show "73. de Broglie Wavelength" '\lambda = \frac{h}{p}'
show "74. Planck-Einstein Relation" 'E = h\nu'
show "75. Time Dilation" '\Delta t^{\prime} = \frac{\Delta t}{\sqrt{1 - \frac{v^2}{c^2}}}'
show "76. Length Contraction" 'L = L_0 \sqrt{1 - \frac{v^2}{c^2}}'
show "77. Relativistic Energy-Momentum" 'E^2 = (pc)^2 + (mc^2)^2'
show "78. Schwarzschild Radius" 'r_s = \frac{2GM}{c^2}'
show "79. Photoelectric Effect" 'E_k = h\nu - \phi'
show "80. Rydberg Formula" '\frac{1}{\lambda} = R\left(\frac{1}{{n_1}^2} - \frac{1}{{n_2}^2}\right)'

# ── Thermodynamics & Statistical Mechanics ──────────────────
section "Thermodynamics"

show "81. Ideal Gas Law" 'PV = nRT'
show "82. Boltzmann Entropy" 'S = k_B \ln \Omega'
show "83. First Law of Thermodynamics" '\Delta U = Q - W'
show "84. Carnot Efficiency" '\eta = 1 - \frac{T_c}{T_h}'
show "85. Stefan-Boltzmann Law" 'P = \sigma A T^4'
show "86. Maxwell-Boltzmann Distribution" 'f(v) = 4\pi n \left(\frac{m}{2\pi kT}\right)^{3/2} v^2 e^{-mv^2/2kT}'
show "87. Planck's Law" 'B(\nu) = \frac{2h\nu^3}{c^2} \cdot \frac{1}{e^{h\nu/kT} - 1}'
show "88. Gibbs Free Energy" 'G = H - TS'
show "89. Clausius Inequality" '\oint \frac{dQ}{T} \leq 0'
show "90. Equipartition Theorem" '\langle E \rangle = \frac{f}{2} k_B T'

# ── Probability & Information ───────────────────────────────
section "Probability and Information"

show "91. Bayes' Theorem" 'P(A \mid B) = \frac{P(B \mid A) \, P(A)}{P(B)}'
show "92. Normal Distribution" 'f(x) = \frac{1}{\sigma\sqrt{2\pi}} e^{-\frac{(x - \mu)^2}{2\sigma^2}}'
show "93. Expected Value" 'E[X] = \sum_{i} x_i \, P(x_i)'
show "94. Variance" '\text{Var}(X) = E[X^2] - (E[X])^2'
show "95. Shannon Entropy" 'H = -\sum_{i} p_i \log_2 p_i'
show "96. Bernoulli Trial" 'P(k) = \frac{n!}{k!(n-k)!} p^k (1-p)^{n-k}'

# ── Famous Constants & Identities ───────────────────────────
section "Famous Constants and Identities"

show "97. Golden Ratio" '\phi = \frac{1 + \sqrt{5}}{2}'
show "98. Euler Product (Riemann Zeta)" '\zeta(s) = \sum_{n=1}^{\infty} \frac{1}{n^s} = \prod_{p} \frac{1}{1 - p^{-s}}'
show "99. Wallis Product" '\frac{\pi}{2} = \prod_{n=1}^{\infty} \frac{4n^2}{4n^2 - 1}'
show "100. Euler's Reflection Formula" '\Gamma(z)\Gamma(1-z) = \frac{\pi}{\sin(\pi z)}'

echo ""
echo "\033[38;2;130;180;255m════════════════════════════════════════════════════════════\033[0m"
echo ""

# Update README.md gallery section
README="$SCRIPT_DIR/README.md"
if [[ -f "$README" ]]; then
  BEGIN='<!-- BEGIN GALLERY -->'
  END='<!-- END GALLERY -->'
  head_part=$(sed "/$BEGIN/q" "$README")
  tail_part=$(sed -n "/$END/,\$p" "$README")
  new_gallery="${head_part}"$'\n\n'"${GALLERY}${tail_part}"
  if [[ "$(cat "$README")" != "$new_gallery" ]]; then
    printf '%s\n' "$new_gallery" > "$README"
    echo "Updated README.md gallery"
  fi
fi
